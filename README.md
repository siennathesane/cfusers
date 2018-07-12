## cfusers

Little worker tool that you can run in Cloud Foundry to manage temporary users. These could be contractors, visitors, temporary teammates, etc.

It's super easy to use. You just have to prepopulate a CSV file with some basics, then push it along with a manifest, and that's it! No more crazy automation, just something nice and simple, easy to use.

What this does:

1. Checks to see if it's time to create a user.
1. Creates a user when it's ready

### User Instructions

* Download this code base.
* Create a file called `prod-manifest.yml` and populate it with the reference manifest. Files starting with `prod*` are ignored by git and cf-cli so it's safe. :)
* Fill out the environment variables. Refer to the reference below for specifics.
* Fill out a CSV file with your users! Feel free to use `temp-users.csv` as a reference template.
* `cf push -f prod-manifest.yml`
* `cf set-health-check cfusers` so it's checked properly.

:warning: The date format matters! :warning:

The date format needs to match this: `2006-01-02T15:04:05Z`. That is January 2nd, 2006 at 3:04pm UTC (Zulu), for reference. This is the date you want your users to be created on.

Below is a quick manifest reference.

```
---
applications:
  - name: cfusers
    no-route: true
    memory: 64M
    disk: 128M
    env:
      GOPACKAGENAME: github.com/mxplusb/cfusers
      # this would be your cloud foundry uaa instance with the uaa:admin:client_credentials user.
      UAA_TARGET:
      UAA_USER:
      UAA_PASSWORD:
      # the cloud controller with uaa:admin:admin_credentials user.
      CAPI_TARGET:
      CAPI_USER:
      CAPI_PASSWORD:
      # how long you want keep users for. syntax reference: https://golang.org/pkg/time/#example_ParseDuration
      # example: you want users to stay for 6 hours and 18 minutes so you would use 6h18m
      USER_KEEPALIVE:
      # since it's a temp user, pick a default password for the users to get.
      DEFAULT_PASSWORD:
      # the name of the CSV file.
      CSV_FILE:
```

If you want to clean things up faster and not wait for users to expire naturally, just change the `USER_KEEPALIVE` variable to `1m` or something like that. The next time it refreshes it's state, it will likely delete all users. Don't remove users from the spreadsheet until they've expired! If you do, this tool won't be able to track them, so resources will be left alone. If you do remove a user by accident, just go through and readd them (anywhere is fine, the order does not matter). This tool will refresh it's user references every 30 seconds, so things happen pretty quickly. If for some reason the app crashes, it'll be okay, it can pick up where it left off. :)

### Dev Instructions

There is a lot of things I want to do, but not many things I have gotten to. To test with random users, just run `dev-reset.py` (all standard library with python3).

If you want to reset the `temp-users.csv` file reference just run `git checkout HEAD -- temp-users.csv`. Please don't check in random users.

### TODO

1. Fix the logging. It's just a bunch of whacky print statements.
1. Wrap this in a web server with a basic GUI. (I'm not good with UIs)
1. Make it extensible. Reading from a CSV file is great but a database would be better.
1. Migrate global cooldowns to local cooldowns. i.e., each user should have it's own expiry.
1. Write tests.
1. Would be cool if this supported k8s.
1. Rewrite this a little more stably. It works as intended, but I crammed this together in like...3 hours with unfamiliar libraries and interfaces. It could be better.

If you like this, I love hard ciders. ;)