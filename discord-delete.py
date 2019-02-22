import aiohttp
import asyncio
import os
import csv

from argparse import ArgumentParser
from csv import DictReader

API = "https://discordapp.com/api/v6"

"""
parser = ArgumentParser()

action = parser.add_subparsers(
	dest="action",
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
"""

class Discord:
	def __init__(self, token):
		self.token = token
		self.session = aiohttp.ClientSession()

	async def close(self):
		await self.session.close()

	async def _req(self, method, endpoint, body=None):
		async with self.session.request(
				method, "{}/{}".format(API, endpoint), 
				headers={ "Authorization": self.token },
				json=body) as resp:
			try:
				data = await resp.json()
			except Exception:
				return None

			# If we're being rate limited, wait for a while.
			if resp.status == 429:
				await asyncio.sleep(data["retry_after"] / 1000)
				return await self._req(method, endpoint, body)

			return data

	async def me(self):
		return await self._req("GET", "/users/@me")

	async def relationships(self):
		return await self._req("GET", "/users/@me/relationships")

	async def relationship_channels(self, body=None):
		return await self._req("POST", "/users/@me/channels", body)

	async def guilds(self):
		return await self._req("GET", "/users/@me/guilds")

	async def guild_messages(self, gid, aid):
		return await self._req("GET", "/guilds/{}/messages/search?author_id={}&include_nsfw=true&limit=25".format(gid, aid))

	async def channels(self):
		return await self._req("GET", "/users/@me/channels")

	async def channel_messages(self, cid, aid):
		return await self._req("GET", "/channels/{}/messages/search?author_id={}&include_nsfw=true&limit=25".format(cid, aid))

	async def delete_message(self, cid, mid):
		return await self._req("DELETE", "/channels/{}/messages/{}".format(cid, mid))

async def delete_from_current(discord):
	"""

	Delete messages from the conversations and servers the user currently participates in.

	"""
	me = await discord.me()

	relations = await discord.relationships()
	for relation in relations:
		await delete_from_relationship(discord, me["id"], relation["id"])

	guilds = await discord.guilds()
	for guild in guilds:
		await delete_from_guild(discord, me["id"], guild["id"])

	channels = await discord.channels()
	for channel in channels:
		await delete_from_channel(discord, me["id"], channel["id"])

async def delete_from_relationship(discord, meid, rid):
	channel = await discord.relationship_channels({
		"recipients": [rid]
	})

	await delete_from_channel(discord, meid, channel["id"])

async def delete_from_guild(discord, meid, gid):
	print("Deleting messages in guild {}.".format(gid))

	messages = await discord.guild_messages(gid, meid)

	if not "total_results" in messages:
		return

	while messages["total_results"] != 0:
		await delete_messages(discord, meid, messages)
		messages = await discord.guild_messages(gid, meid)

async def delete_from_channel(discord, meid, cid):
	print("Deleting messages in channel {}.".format(cid))

	messages = await discord.channel_messages(cid, meid)

	if not "total_results" in messages:
		return

	while messages["total_results"] != 0:
		await delete_messages(discord, meid, messages)
		messages = await discord.channel_messages(cid, meid)

async def delete_messages(discord, meid, messages):
	for context in messages["messages"]:
		for message in context:
			print("Deleted message {}.".format(message["id"]))
			await discord.delete_message(message["channel_id"], message["id"])

async def delete_from_all(discord):
	"""

	Delete messages from the user's entire history by using a data download package.

	"""
	if not os.path.isdir("./messages"):
		return

	for filename in os.listdir("./messages"):
		if not os.path.isdir("./messages/{}".format(filename)):
			return

		messages = open("./messages/{}/messages.csv".format(filename), "r")
		reader = DictReader(messages)

		for line in reader:
			await discord.delete_message(filename, line["ID"])

async def main():
	token = os.environ.get("DISCORD_TOKEN")

	if token:
		discord = Discord(token)

		await delete_from_current(discord)
		# await delete_from_all(discord)
		await discord.close()
	else:
	   print("You must specify a Discord auth token by setting DISCORD_TOKEN.")

loop = asyncio.get_event_loop()
loop.run_until_complete(main())
