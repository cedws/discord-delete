import aiohttp
import asyncio
import os
import csv

from aiohttp import ClientResponseError
from csv import DictReader

API = "https://discordapp.com/api/v6"

class Discord:
    def __init__(self, token):
        self.token = token
        self.session = aiohttp.ClientSession()

    async def close(self):
        await self.session.close()

    async def _req(self, method, endpoint, body=None):
        url = "{}/{}".format(API, endpoint)
        headers = { "Authorization": self.token }

        async with self.session.request(method, url,
                headers=headers, json=body) as resp:
            try:
                json = await resp.json()
            except ClientResponseError:
                # If we're being rate limited, wait for a while.
                if resp.status == 429:
                    await asyncio.sleep(json.get("retry_after") / 1000)
                    return await self._req(method, endpoint, body)

                return {}
            except Exception:
                raise

            return json

    async def me(self):
        return await self._req("GET", "/users/@me")

    async def relationships(self):
        return await self._req("GET", "/users/@me/relationships")

    async def relationship_channels(self, rid):
        return await self._req("POST", "/users/@me/channels", {
            "recipients": [rid]
        })

    async def guilds(self):
        return await self._req("GET", "/users/@me/guilds")

    async def guild_msgs(self, gid, aid):
        return await self._req("GET", "/guilds/{}/messages/search?author_id={}&include_nsfw=true&limit=25".format(gid, aid))

    async def channels(self):
        return await self._req("GET", "/users/@me/channels")

    async def channel_msgs(self, cid, aid):
        return await self._req("GET", "/channels/{}/messages/search?author_id={}&include_nsfw=true&limit=25".format(cid, aid))

    async def delete_msg(self, cid, mid):
        return await self._req("DELETE", "/channels/{}/messages/{}".format(cid, mid))

async def delete_from_current(discord):
    """

    Delete messages from the conversations and servers the user currently participates in.

    """
    meid = (await discord.me()).get("id")

    channels = await discord.channels()
    cids = [channel.get("id") for channel in channels]

    for cid in cids:
        await delete_from_channel(discord, meid, cid)

    relations = await discord.relationships()
    rids = [relation.get("id") for relation in relations]

    for rid in rids:
        channel = await discord.relationship_channels(rid)

        if not channel.get("id") in cids:
            await delete_from_channel(discord, meid, channel.get("id"))

    guilds = await discord.guilds()
    gids = [guild.get("id") for guild in guilds]

    for gid in gids:
        await delete_from_guild(discord, meid, gid)

async def delete_from_channel(discord, meid, cid):
    print("Deleting messages in channel {}...".format(cid))

    results = True
    while results:
        mlist = await discord.channel_msgs(cid, meid)
        await delete_msgs(discord, meid, mlist.get("messages"))

        results = mlist.get("total_results")

async def delete_from_guild(discord, meid, gid):
    print("Deleting messages in guild {}...".format(gid))
    
    results = True
    while results:
        mlist = await discord.guild_msgs(gid, meid)
        await delete_msgs(discord, meid, mlist.get("messages"))

        results = mlist.get("total_results")

async def delete_msgs(discord, meid, msgs):
    for context in msgs:
        for msg in context:
            if not msg.get("hit"):
                continue

            print("\t- {}".format(msg.get("id")))
            await discord.delete_msg(
                msg.get("channel_id"), 
                msg.get("id")
            )

async def delete_from_all(discord):
    """

    Delete messages from the user's entire history by using a data download package.

    """
    if not os.path.isdir("./messages"):
        return

    for filename in os.listdir("./messages"):
        if not os.path.isdir("./messages/{}".format(filename)):
            return

        msgs = open("./messages/{}/messages.csv".format(filename), "r")
        reader = DictReader(msgs)

        for line in reader:
            await discord.delete_msg(filename, line["ID"])

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