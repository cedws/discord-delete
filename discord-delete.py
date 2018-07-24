import time
import argparse
import requests
import os
import csv

# Constants
API = "https://discordapp.com/api/v6" 

parser = argparse.ArgumentParser()

action = parser.add_subparsers(
	dest="action"
)

wipe = action.add_parser(
	"wipe", 
	help="Wipe messages from all channels you have participated in. Requires a data download package.")

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
	required=True,
	help="Your Discord account email address."
)

parser.add_argument(
	"--password", 
	required=True,
	help="Your Discord account password."
)

args = parser.parse_args()

def main():
	discord = Discord(args.email, args.password)

	if clear:
		clear_messages(discord)

	# TODO: Clean this up and move it into the `wipe` action. 
	"""
	for filename in os.listdir("./messages"):
		if not os.path.isdir("./messages/{}".format(filename)): 
			return

		messages = open("./messages/{}/messages.csv".format(filename), "r")
		reader = csv.DictReader(messages)

		for line in reader:
			req = requests.delete(
				"https://discordapp.com/api/v6/channels/{}/messages/{}".format(filename, line["ID"]), 
				headers=headers
			)
			print(req.url)
			print(req.text)
			time.sleep(0.2)
			"""

def clear_messages(discord):
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