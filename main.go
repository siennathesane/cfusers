package main

import (
	"time"
	"os"
	log "github.com/sirupsen/logrus"
	"github.com/gocarina/gocsv"
	"github.com/cloudfoundry-community/go-cfclient"
	"fmt"
	"github.com/cloudfoundry-community/go-uaa"
	"strings"
)

var (
	uaaTarget        = os.Getenv("UAA_TARGET")
	uaaUser          = os.Getenv("UAA_USER")
	uaaPassword      = os.Getenv("UAA_PASSWORD")
	capiTarget       = os.Getenv("CAPI_TARGET")
	capiUser         = os.Getenv("CAPI_USER")
	capiPassword     = os.Getenv("CAPI_PASSWORD")
	userKeepAlive    = os.Getenv("USER_KEEPALIVE")
	baselinePassword = os.Getenv("DEFAULT_PASSWORD")
)

// User defines the CSV file format.
type User struct {
	GivenName  string `csv:"FirstName"`
	FamilyName string `csv:"LastName"`
	Email      string `csv:"Email"`
	// This needs to be in RFC 822 format. Reference: 2006-01-02T15:04:05Z07:00
	DateStart string `csv:"DateStart"`
}

func main() {
	fmt.Println("bootstrapping.")
	c := &cfclient.Config{
		ApiAddress: capiTarget,
		Username:   capiUser,
		Password:   capiPassword,
	}
	cfClient, err := cfclient.NewClient(c)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("connected to cf.")

	uaaClient, err := uaa.NewWithClientCredentials(uaaTarget, "", uaaUser, uaaPassword, uaa.OpaqueToken)
	if err != nil {
		log.Fatalf("error connecting to uaa. %s", err)
	}
	// enable this for debugging.
	// uaaClient.Verbose = true
	fmt.Println("connected to cf-uaa.")

	users := marshallUsers("temp-users.csv")
	fmt.Println("loaded reference file.")

	validateLifecycle(cfClient, uaaClient, users)
}

func marshallUsers(fn string) []*User {
	users := []*User{}
	custFile, err := os.OpenFile(fn, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Fatalf("error opening file. %s", err)
	}
	if err := gocsv.UnmarshalFile(custFile, &users); err != nil {
		log.Fatalf("error marshalling customer file. %s", err)
	}
	custFile.Close()
	return users
}

func validateLifecycle(c *cfclient.Client, a *uaa.API, u []*User) {
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()

	keepAliveLength, err := time.ParseDuration(userKeepAlive)
	if err != nil {
		log.Fatalf("cannot parse user keepalive time. %s", err)
	}

	for {
		select {
		case _ = <-ticker.C:
			fmt.Println("ticking.")
			for _, user := range u {
				// this means that the users exist in the spreadsheet but are essentially unmanaged at this point.
				if user.DateStart == "" {
					fmt.Printf("skipping user %s since they don't have an assigned date.\n", user.Email)
					continue
				}
				now := time.Now().UTC()
				startTime, err := time.Parse(time.RFC3339, user.DateStart)
				if err != nil {
					log.Errorf("error parsing %s start date. %s\n", user.Email, err)
				}

				// check to make sure the user exists.
				exists, err := userExists(c, user)
				if err != nil {
					log.Error(err)
				}

				expiryTime := startTime.Add(keepAliveLength)

				// if the user exists, check to make sure the haven't expired.
				if exists {
					// if the user has expired, delete them.
					if expiryTime.Before(now) {
						fmt.Printf("deleting user %s since their access has expired.\n", user.Email)
						deleteUser(a, c, user)
						continue
					}

					// validate their org exists. this is really just to prevent things from getting out of whack.
					orgExists, err := orgExists(c, user)
					if err != nil {
						log.Error(err)
					}
					if !orgExists {
						buildOrg(a, c, user)
					}
				}

				// if the user does not exist
				if !exists {
					// and they have already expired.
					if expiryTime.Before(now) {
						// skip
						continue
					}
					fmt.Printf("user %s is due to be created by %s\n", user.Email, startTime.String())
					// create them.
					if startTime.Before(now) {
						fmt.Printf("creating user %s and their associated org and space.\n", user.Email)
						buildUser(a, c, user)
					}
				}
			}}
	}
}

func buildUser(a *uaa.API, c *cfclient.Client, u *User) {
	// create the user in UAA.
	userRef := uaa.User{
		Username: usernameShortener(u),
		Password: baselinePassword,
		Emails: []uaa.Email{{
			Value:   u.Email,
			Primary: func() *bool { b := true; return &b }(),
		}},
		Name: &uaa.UserName{
			FamilyName: u.FamilyName,
			GivenName:  u.GivenName,
		},
	}
	//fmt.Printf("creating user as such: %#v\n", userRef)
	user, err := a.CreateUser(userRef)
	if err != nil {
		log.Errorf("error creating %s user. %s", usernameShortener(u), err)
	}
	fmt.Printf("created %s user.\n", usernameShortener(u))

	org, err := c.CreateOrg(cfclient.OrgRequest{
		Name: fmt.Sprintf("%s-org", usernameShortener(u)),
	})
	if err != nil {
		log.Errorf("error creating %s-org. %s", usernameShortener(u), err)
	}
	fmt.Printf("created %s-org.\n", usernameShortener(u))

	_, err = c.AssociateOrgManager(org.Guid, user.ID)
	if err != nil {
		log.Errorf("error associating %s with %s-org. %s", usernameShortener(u), usernameShortener(u), err)
	}
	fmt.Printf("associated %s with %s-org as OrgManager.\n", usernameShortener(u), usernameShortener(u))

	_, err = c.AssociateOrgUser(org.Guid, user.ID)
	if err != nil {
		log.Errorf("error associating %s with %s-org as org user. %s", usernameShortener(u), usernameShortener(u), err)
	}
	fmt.Printf("associated %s with %s-org as OrgUser.", usernameShortener(u), usernameShortener(u))

	_, err = c.CreateSpace(cfclient.SpaceRequest{
		Name:             fmt.Sprintf("%s-dev", usernameShortener(u)),
		OrganizationGuid: org.Guid,
		ManagerGuid:      []string{user.ID},
		DeveloperGuid:    []string{user.ID},
	})
	if err != nil {
		log.Errorf("error creating %s-dev space. %s", usernameShortener(u), err)
	}
	fmt.Printf("associated %s with %s-dev as SpaceManager and SpaceDeveloper.\n", usernameShortener(u), usernameShortener(u))
	return
}

func usernameShortener(u *User) string {
	return fmt.Sprintf("%s%s", strings.ToLower(string([]rune(u.GivenName)[0])), strings.ToLower(u.FamilyName))
}

func buildOrg(a *uaa.API, c *cfclient.Client, u *User) {
	// get our user
	user, err := a.GetUserByUsername(usernameShortener(u), "", "")
	if err != nil {
		log.Errorf("error getting %s user. %s", u.Email, err)
	}
	fmt.Printf("got %s user.\n", u.Email)

	org, err := c.CreateOrg(cfclient.OrgRequest{
		Name: fmt.Sprintf("%s-org", usernameShortener(u)),
	})
	if err != nil {
		log.Errorf("error creating %s-org. %s", usernameShortener(u), err)
	}
	fmt.Printf("created %s-org.\n", usernameShortener(u))

	_, err = c.AssociateOrgManager(org.Guid, user.ID)
	if err != nil {
		log.Errorf("error associating %s with %s-org. %s", user.Emails[0].Value, usernameShortener(u), err)
	}
	fmt.Printf("associated %s with %s-org as OrgManager.\n", usernameShortener(u), usernameShortener(u))

	_, err = c.AssociateOrgUser(org.Guid, user.ID)
	if err != nil {
		log.Errorf("error associating %s with %s-org as org user. %s", usernameShortener(u), usernameShortener(u), err)
	}
	fmt.Printf("associated %s with %s-org as OrgUser.\n", usernameShortener(u), usernameShortener(u))

	_, err = c.CreateSpace(cfclient.SpaceRequest{
		Name:             fmt.Sprintf("%s-dev", usernameShortener(u)),
		OrganizationGuid: org.Guid,
		ManagerGuid:      []string{user.ID},
		DeveloperGuid:    []string{user.ID},
	})
	if err != nil {
		log.Errorf("error creating %s-dev space. %s", usernameShortener(u), err)
	}
	fmt.Printf("associated %s with %s-dev as SpaceManager and SpaceDeveloper.\n", usernameShortener(u), usernameShortener(u))
	return
}

// wipe a user from cf then uaa.
func deleteUser(a *uaa.API, c *cfclient.Client, u *User) {
	preDeleteOrgRef, err := c.GetOrgByName(fmt.Sprintf("%s-org", usernameShortener(u)))
	if err != nil {
		log.Error(err)
	}
	err = c.DeleteOrg(preDeleteOrgRef.Guid, true, false)
	if err != nil {
		log.Errorf("can't delete %s-org. %s", err)
	}
	testUser, err := a.GetUserByUsername(usernameShortener(u), "", "")
	if err != nil {
		log.Errorf("error getting %s to delete. %s\n", usernameShortener(u), err)
	}
	_, err = a.DeleteUser(testUser.ID)
	if err != nil {
		log.Errorf("error deleting %s user. %s", usernameShortener(u), err)
	}
	fmt.Printf("successfully deleted %s from cf and cf-uaa.\n", usernameShortener(u))
}

func userExists(c *cfclient.Client, u *User) (bool, error) {
	users, err := c.ListUsers()
	if err != nil {
		return false, err
	}
	userRef := users.GetUserByUsername(usernameShortener(u))
	if userRef.Guid == "" {
		fmt.Printf("user %s does not exist.\n", usernameShortener(u))
		return false, nil
	} else {
		fmt.Printf("user %s exists.\n", usernameShortener(u))
		return true, nil
	}
}

func orgExists(c *cfclient.Client, u *User) (bool, error) {
	targetOrg := fmt.Sprintf("%s-org", usernameShortener(u))
	org, err := c.GetOrgByName(targetOrg)
	if err != nil {
		return false, err
	}
	if org.Name == targetOrg {
		fmt.Printf("found %s.\n", targetOrg)
		return true, nil
	} else {
		fmt.Printf("missing %s.\n", targetOrg)
		return false, nil
	}
}
