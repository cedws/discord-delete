import time
import os
import csv
import argparse
import requests

# Constants
API = "https://discordapp.com/api/v6" 

parser = argparse.ArgumentParser()

action = parser.add_subparsers(
    dest="action"
)

wipe = action.add_parser(
    "wipe", 
    help="Wipe messages from all channels you have participated in. Requires a data download package."
)

wipe.add_argument(
    "data",
    help="Root directory of the uncompressed data archive."
)

clear = action.add_parser(
    "clear", 
    help="Clear messages from channels you are currently participating in."
)

parser.add_argument(
    "--paranoid", 
    action="store_true", 
    help="Overwrite messages with blank text before deleting."
)

parser.add_argument(
    "--email", 
    "-e",
    required=True,
    help="Your Discord account email address."
)

parser.add_argument(
    "--password", 
    "-p",
    required=True,
    help="Your Discord account password."
)

args = parser.parse_args()

def main():
    discord = Discord(args.email, args.password)

    if wipe:
        wipe_messages(discord, args.data)

    if clear:
        clear_messages(discord)

def wipe_messages(discord, root):
    path = os.path.join(root, "messages")

    for sub in os.listdir(path):
        if not os.path.isdir("{}/{}".format(path, sub)): 
            continue
        
        messages = open("{}/{}/messages.csv".format(path, sub), "r")
        reader = csv.DictReader(messages)
        
        for line in reader:
            print(discord.delete_message(filename, line["ID"]))

def clear_messages(discord):
    # TODO: Delete from guilds too.
    channels = discord.get_channels()
    
    for channel in channels:
        messages = discord.get_messages(channel["id"])

        for message in messages:
            # TODO: Better output.
            print(discord.delete_message(channel["id"], message["id"]))

class Discord:
    def __init__(self, email, password):
        self.token = Discord.__token(email, password)

    def __token(email, password):
        return requests.post("{}/auth/login".format(API), json={ 
            "email": email,
            "password": password
        }).json()["token"]

    def __get(self, endpoint):
        res = requests.get(
            endpoint, 
            headers={ "Authorization": self.token }
        )

        data = res.json()

        # If we're being rate limited, wait for a while.
        if res.status_code is 429:
            time.sleep(data["retry_after"])
            return self.__get(endpoint)

        return data

    def __delete(self, endpoint):
        res = requests.delete(
            endpoint, 
            headers={ "Authorization": self.token }
        )

        data = res.json()

        # If we're being rate limited, wait for a while.
        if res.status_code is 429:
            time.sleep(data["retry_after"])
            return self.__get(endpoint)

        return data
    
    def get_me(self):
        return self.__get("{}/users/@me".format(API))

    def get_channels(self):
        return self.__get("{}/users/@me/channels".format(API))

    def get_guilds(self):
        return self.__get("{}/users/@me/guilds".format(API))

    def get_messages(self, cid):
        return self.__get("{}/channels/{}/messages".format(API, cid))

    def delete_message(self, cid, mid):
        return self.__delete("{}/channels/{}/messages/{}".format(API, cid, mid))

if __name__ == "__main__":
    main()