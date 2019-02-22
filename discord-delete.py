import aiohttp
import asyncio
import os
import csv

from csv import DictReader

API = "https://discordapp.com/api/v6"

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
    meid = (await discord.me())["id"]

    channels = await discord.channels()
    cids = [channel["id"] for channel in channels]
    for cid in cids:
        await delete_from_channel(discord, meid, cid)

    relations = await discord.relationships()
    rids = [relation["id"] for relation in relations]
    for rid in rids:
        channel = await discord.relationship_channels({
            "recipients": [rid]
        })

        if not channel["id"] in cids:
            await delete_from_channel(discord, meid, channel["id"])

    guilds = await discord.guilds()
    gids = [guild["id"] for guild in guilds]
    for gid in gids:
        await delete_from_guild(discord, meid, gid)

async def delete_from_channel(discord, meid, cid):
    print("Deleting messages in channel {}...".format(cid))

    messages = await discord.channel_messages(cid, meid)

    while messages.get("total_results"):
        await delete_messages(discord, meid, messages)
        messages = await discord.channel_messages(cid, meid)

async def delete_from_guild(discord, meid, gid):
    print("Deleting messages in guild {}...".format(gid))

    messages = await discord.guild_messages(gid, meid)

    while messages.get("total_results"):
        await delete_messages(discord, meid, messages)
        messages = await discord.guild_messages(gid, meid)

async def delete_messages(discord, meid, messages):
    for context in messages["messages"]:
        for message in context:
            if not message.get("hit"):
                continue

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
        await delete_from_all(discord)

        await discord.close()
    else:
       print("You must specify a Discord auth token by setting DISCORD_TOKEN.")

loop = asyncio.get_event_loop()
loop.run_until_complete(main())
