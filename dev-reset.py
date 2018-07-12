#! /usr/bin/env python3
import csv
import datetime
import random
import string
import sys
from collections import OrderedDict
from os import path


class DevReset(object):
    def __init__(self, where, amount):
        """
        reset our dev environment.
        :param where:
        :param amount:
        """
        self.vowels = "aeiou"
        self.consonants = "".join(set(string.ascii_lowercase) - set(self.vowels))
        self.now = datetime.datetime.utcnow()
        self.temp_users = []
        self.amount = amount
        self.where = where

    def generate_word(self, length):
        """
        generate some words.
        :param length: how long do you want this word to be?
        :return:
        """
        word = ""
        for i in range(length):
            if i % 2 == 0:
                word += random.choice(self.consonants)
            else:
                word += random.choice(self.vowels)
        return word

    def generate_randos(self):
        """
        create some weirdos.
        """
        for i in range(self.amount):
            self.temp_users.append(
                OrderedDict([
                    ("FirstName", self.generate_word(random.randint(5, 15))),
                    ("LastName", self.generate_word(random.randint(5, 25))),
                    ("Email", "{0}@{1}.com".format(self.generate_word(random.randint(4, 10)), self.generate_word(random.randint(5, 10)))),
                    ("DateStart", self.now.strftime("%Y-%m-%dT%H:%M:%SZ"))
                ])
            )

    def read(self):
        """
        where are we reading from?
        """
        with open(self.where, 'r') as fh:
            reader = csv.DictReader(fh)
            [self.temp_users.append(i) for i in reader]

    def write(self):
        """
        where are we writing to?
        """
        with open(self.where, 'w') as fh:
            field_names = ["FirstName", "LastName", "Email", "DateStart"]
            writer = csv.DictWriter(fh, fieldnames=field_names)
            writer.writeheader()
            writer.writerows(self.temp_users)


def main():
    if len(sys.argv) == 1:
        print("you need to tell me how many randoms to create.")
        sys.exit(1)
    try:
        rando_count = int(sys.argv[1])
    except TypeError:
        print("gotta tell me (as an integer) how many random users to create, yo.")
        sys.exit(1)
    dev = DevReset(where=path.join(path.abspath("."), "temp-users.csv"), amount=rando_count)
    dev.read()
    dev.generate_randos()
    dev.write()
    print("you may now cf push.")


if __name__ == "__main__":
    main()
