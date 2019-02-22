import aiohttp
import asyncio
import os
import csv

from aiohttp import ClientResponseError
from csv import DictReader

API = "https://discordapp.com/api/v6"
ENDPOINTS = {
    "me":               "/users/@me",
    "relationships":    "/users/@me/relationships",
    "channels":         "/users/@me/channels",
    "guilds":           "/users/@me/guilds",
    "guild_msgs":       (
                            "/guilds/{}/messages/search"
                            "?author_id={}"
                            "&include_nsfw=true"
                            "&limit=25"
                        ),
    "channels":         "/users/@me/channels",
    "channel_msgs":     (
                            "/channels/{}/messages/search"
                            "?author_id={}"
                            "&include_nsfw=true"
                            "&limit=25"
                        ),
    "delete_msg":       "/channels/{}/messages/{}"
}

class Discord:
    def __init__(self, token):
        self.token = token
        self.session = aiohttp.ClientSession(loop=loop)

    async def __aenter__(self):
        return self

    async def __aexit__(self, *args):
        await self.session.close()

    async def _req(self, method, endpoint, body=None):
        url = "{}/{}".format(API, endpoint)
        headers = { "Authorization": self.token }

        async with self.session.request(method, url,
                headers=headers,
                json=body) as resp:
            try:
                json = await resp.json()
            except ClientResponseError:
                # If we're being rate limited, wait for a while.
                if resp.status == 429:
                    assert "retry_after" in json
                    await asyncio.sleep(json.get("retry_after") / 1000)
                    return await self._req(method, endpoint, body)
                if resp.status >= 500:
                    exit("Server reported internal error.")
                return {}
            except Exception:
                raise

            return json

    async def me(self):
        return await self._req("GET",
            ENDPOINTS["me"])

    async def relationships(self):
        return await self._req("GET",
            ENDPOINTS["relationships"])

    async def relationship_channels(self, r_id):
        return await self._req("POST",
            ENDPOINTS["channels"], {
                "recipients": [r_id]
            })

    async def guilds(self):
        return await self._req("GET",
            ENDPOINTS["guilds"])

    async def guild_msgs(self, g_id, a_id):
        return await self._req("GET",
            ENDPOINTS["guild_msgs"].format(g_id, a_id))

    async def channels(self):
        return await self._req("GET",
            ENDPOINTS["channels"])

    async def channel_msgs(self, c_id, a_id):
        return await self._req("GET",
            ENDPOINTS["channel_msgs"].format(c_id, a_id))

    async def delete_msg(self, c_id, m_id):
        return await self._req("DELETE",
            ENDPOINTS["delete_msg"].format(c_id, m_id))

    async def delete_from_current(self):
        """

        Delete messages from the conversations and servers the user currently participates in.

        """
        me_id = (await self.me()).get("id")
        assert me_id

        channels = await self.channels()
        c_ids = [channel.get("id") for channel in channels]

        for c_id in c_ids:
            await self.delete_from_channel(me_id, c_id)

        relations = await self.relationships()
        r_ids = [relation.get("id") for relation in relations]

        for r_id in r_ids:
            channel = await self.relationship_channels(r_id)

            if not channel.get("id") in c_ids:
                await self.delete_from_channel(me_id, channel.get("id"))

        guilds = await self.guilds()
        g_ids = [guild.get("id") for guild in guilds]

        for g_id in g_ids:
            await self.delete_from_guild(me_id, g_id)

    async def delete_from_channel(self, me_id, c_id):
        print("Deleting messages in channel {}...".format(c_id))

        results = True
        while results:
            mlist = await self.channel_msgs(c_id, me_id)
            await self.delete_msgs(mlist.get("messages"))

            results = mlist.get("total_results")

    async def delete_from_guild(self, me_id, g_id):
        print("Deleting messages in guild {}...".format(g_id))

        results = True
        while results:
            mlist = await self.guild_msgs(g_id, me_id)
            await self.delete_msgs(mlist.get("messages"))

            results = mlist.get("total_results")

    async def delete_msgs(self, msgs):
        if msgs == None:
            return

        for context in msgs:
            msg = next(m for m in context if m.get("hit"))
            assert msg

            print("\t- {}".format(msg.get("id")))
            await self.delete_msg(
                msg.get("channel_id"),
                msg.get("id")
            )

    async def delete_from_all(self):
        """

        Delete messages from the user's entire history by using a data download package.

        """

        # TODO: Clean this up.
        if not os.path.isdir("./messages"):
            return

        for filename in os.listdir("./messages"):
            if not os.path.isdir("./messages/{}".format(filename)):
                return

            msgs = open("./messages/{}/messages.csv".format(filename), "r")
            reader = DictReader(msgs)

            for line in reader:
                await self.delete_msg(filename, line["ID"])

async def main():
    # TODO: More secure way of passing this value.
    token = os.environ.get("DISCORD_TOKEN")

    if token:
        async with Discord(token) as client:
            await client.delete_from_current()
            await client.delete_from_all()
    else:
       print("You must pass a Discord auth token by setting DISCORD_TOKEN.")

loop = asyncio.get_event_loop()
loop.run_until_complete(main())