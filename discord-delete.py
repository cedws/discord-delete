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
"""

class Discord:
    async def login(email, password):
        payload = {
            "email": email,
            "password": password
        }

        async with aiohttp.ClientSession() as session:
            async with session.post("{}/auth/login".format(API), json=payload) as resp:
                return Discord((await resp.json())["token"])

    def __init__(self, token):
        self.token = token

    async def _req(self, method, endpoint):
        async with aiohttp.ClientSession() as session:
            async with session.request(method, "{}/{}".format(API, endpoint), headers={ "Authorization": self.token }) as resp:
                try:
                    data = await resp.json()
                except Exception:
                    return None

                # If we're being rate limited, wait for a while.
                if resp.status == 429:
                    await asyncio.sleep(data["retry_after"] / 1000)
                    return await self._req(method, endpoint)

                return data

    async def me(self):
        return await self._req("GET", "/users/@me")

    async def channels(self):
        return await self._req("GET", "/users/@me/channels")

    async def guilds(self):
        return await self._req("GET", "/users/@me/guilds")

    async def guild_messages(self, gid, aid):
        return await self._req("GET", "/guilds/{}/messages/search?author_id={}&include_nsfw=true&limit=25".format(gid, aid))

    async def channel_messages(self, cid, aid):
        return await self._req("GET", "/channels/{}/messages/search?author_id={}&include_nsfw=true&limit=25".format(cid, aid))

    async def delete_message(self, cid, mid):
        return await self._req("DELETE", "/channels/{}/messages/{}".format(cid, mid))

async def delete_from_current(discord):
    me = await discord.me()
    guilds = await discord.guilds()

    for guild in guilds:
        messages = await discord.guild_messages(guild["id"], me["id"])

        if not "total_results" in messages:
            continue

        while messages["total_results"] != 0:
            for context in messages["messages"]:
                for message in context:
                    if message["author"]["id"] == me["id"]:
                        print(message["id"])
                        await discord.delete_message(message["channel_id"], message["id"])

            messages = await discord.guild_messages(guild["id"], me["id"])

    channels = await discord.channels()

    for channel in channels:
        messages = await discord.channel_messages(channel["id"], me["id"])

        if not "messages" in messages:
            continue

        for message in messages["messages"]:
            for context in messages["messages"]:
                for message in context:
                    if message["author"]["id"] == me["id"]:
                        print(message["id"])
                        await discord.delete_message(message["channel_id"], message["id"])

            messages = await discord.channel_messages(channel["id"], me["id"])

async def delete_from_all(discord):
    for filename in os.listdir("./messages"):
        if not os.path.isdir("./messages/{}".format(filename)):
            return

        messages = open("./messages/{}/messages.csv".format(filename), "r")
        reader = DictReader(messages)

        for line in reader:
            print(line["ID"])
            await discord.delete_message(filename, line["ID"])

async def main():
    discord = await Discord.login("", "")
    await delete_from_current(discord)
    # await delete_from_all(discord)

loop = asyncio.get_event_loop()
loop.run_until_complete(main())
