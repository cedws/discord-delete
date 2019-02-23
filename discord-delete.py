import aiohttp
import asyncio
import os
import csv

from csv import DictReader

API = "https://discordapp.com/api/v6"
LIMIT = 25
ENDPOINTS = {
    "me":               "/users/@me",
    "relationships":    "/users/@me/relationships",
    "channels":         "/users/@me/channels",
    "guilds":           "/users/@me/guilds",
    "guild_msgs":       (
                            "/guilds/{}/messages/search"
                            "?author_id={}"
                            "&include_nsfw=true"
                            "&limit={}"
                        ),
    "channels":         "/users/@me/channels",
    "channel_msgs":     (
                            "/channels/{}/messages/search"
                            "?author_id={}"
                            "&include_nsfw=true"
                            "&limit={}"
                        ),
    "delete_msg":       "/channels/{}/messages/{}"
}

class CacheItem:
    def __init__(self, method, endpoint, json, body=None):
        self.method = method
        self.endpoint = endpoint
        self.body = body
        self.json = json

class Cache:
    def __init__(self):
        self.cached = []

    def get(self, method, endpoint, body=None):
        return next(
            (i for i in self.cached if
                i.method == method and
                i.endpoint == endpoint and
                i.body == body
            ),
            None
        )

    def put(self, method, endpoint, json, body=None):
        item = CacheItem(method, endpoint, json, body)
        self.cached.append(item)

class Discord:
    def __init__(self, token):
        self.token = token
        self.session = aiohttp.ClientSession(loop=loop)
        self.cache = Cache()

    async def __aenter__(self):
        return self

    async def __aexit__(self, *args):
        await self.session.close()

    async def _req(self, method, endpoint, body=None, cache=False):
        if cache:
            cached = self.cache.get(method, endpoint, body=body)
            # We've made a request like this already and it's safe to return the
            # cached response.
            if cached:
                return cached.json

        url = "{}/{}".format(API, endpoint)
        headers = { "Authorization": self.token }

        async with self.session.request(method, url,
                headers=headers,
                json=body) as resp:
            # Internal Server Error
            if resp.status >= 500:
                exit("Server reported internal error.")
            # Not Found
            if resp.status == 404:
                return {}
            # No Content
            if resp.status == 204:
                return {}

            # We should definitely get a JSON response if the status wasn't
            # one of the above.
            json = await resp.json()

            # Too Many Requests
            if resp.status == 429:
                # If we're being rate limited, wait for a while.
                assert "retry_after" in json
                await asyncio.sleep(json.get("retry_after") / 1000)
                return await self._req(method, endpoint, body)
            # OK
            if resp.status == 200:
                if cache:
                    self.cache.put(method, endpoint, json, body=body)

                return json

            # Return an empty response, something unusual happened. Won't be
            # cached.
            return {}

    async def me(self):
        return await self._req(
            "GET",
            ENDPOINTS["me"],
            cache=True
        )

    async def relationships(self):
        return await self._req(
            "GET",
            ENDPOINTS["relationships"],
            cache=True,
        )

    async def relationship_channels(self, r_id):
        return await self._req(
            "POST",
            ENDPOINTS["channels"],
            body={ "recipients": [r_id] },
            cache=True,
        )

    async def guilds(self):
        return await self._req(
            "GET",
            ENDPOINTS["guilds"],
            cache=True
        )

    async def guild_msgs(self, g_id, a_id):
        return await self._req(
            "GET",
            ENDPOINTS["guild_msgs"].format(g_id, a_id, LIMIT)
        )

    async def channels(self):
        return await self._req(
            "GET",
            ENDPOINTS["channels"],
            cache=True
        )

    async def channel_msgs(self, c_id, a_id):
        return await self._req(
            "GET",
            ENDPOINTS["channel_msgs"].format(c_id, a_id, LIMIT)
        )

    async def delete_msg(self, c_id, m_id):
        return await self._req(
            "DELETE",
            ENDPOINTS["delete_msg"].format(c_id, m_id),
            cache=True,
        )

    async def delete_from_current(self):
        """

        Delete messages from the conversations and servers the user currently
        participates in.

        """
        # Request information about the user and get their unique ID.
        # The user should always have an ID.
        me_id = (await self.me()).get("id")
        assert me_id

        # Gather a list of channels that the user is in. Channels are direct
        # messages which may be from groups or 1-to-1. Direct message
        # channels that have been hidden are not returned by the API.
        # We know the API keeps hidden channels open since it returns the
        # same channel ID if the channel is reopened.
        channels = await self.channels()
        c_ids = [c.get("id") for c in channels]

        for c_id in c_ids:
            await self.delete_from_channel(me_id, c_id)

        # Gather a lot of relationships the user has. These are people that the
        # user has in one of their lists (All/Pending/Blocked).
        relations = await self.relationships()
        r_ids = [r.get("id") for r in relations]

        # This code gathers a list of recipients who the user is in contact with,
        # but filters only those who are the *sole* recipient (that the
        # user might be friends with).
        c_recipients = [
            c.get("recipients")[0]
            for c in channels
            if len(c.get("recipients")) == 1
        ]
        c_recipient_ids = [c_r.get("id") for c_r in c_recipients]

        for r_id in r_ids:
            # Avoid making unnecessary requests by seeing if we already
            # checked the recipient channel earlier. Checking relationship
            # channels requires double the requests so we need to avoid doing it
            # if possible.
            if not r_id in c_recipient_ids:
                channel = await self.relationship_channels(r_id)
                await self.delete_from_channel(me_id, channel.get("id"))

        # Gather a list of guilds (known by many as "servers"). This doesn't
        # include guilds that the user has left.
        guilds = await self.guilds()
        g_ids = [g.get("id") for g in guilds]

        for g_id in g_ids:
            await self.delete_from_guild(me_id, g_id)

    async def delete_from_channel(self, me_id, c_id):
        print("Deleting messages in channel {}...".format(c_id))

        # Avoids making unnecessary requests if the number of messages returned
        # is less that LIMIT, indicating there isn't another page of messages.
        results = LIMIT
        while results >= LIMIT:
            mlist = await self.channel_msgs(c_id, me_id)
            messages = mlist.get("messages")

            await self.delete_msgs(messages)
            # We avoid using the "total_results" field as it is often inaccurate.
            results = len(messages)

    async def delete_from_guild(self, me_id, g_id):
        print("Deleting messages in guild {}...".format(g_id))

        # Avoids making unnecessary requests if the number of messages returned
        # is less that LIMIT, indicating there isn't another page of messages.
        results = LIMIT
        while results >= LIMIT:
            mlist = await self.guild_msgs(g_id, me_id)
            messages = mlist.get("messages")

            await self.delete_msgs(messages)
            # We avoid using the "total_results" field as it is often inaccurate.
            results = len(messages)

    async def delete_msgs(self, msgs):
        if msgs == None:
            return

        for context in msgs:
            # The "hit" field is highlighted message. The other messages are for
            # context and may not be authored by the user, so we ignore them.
            # At least one message in the context group should have the "hit"
            # field.
            msg = next(m for m in context if m.get("hit"))
            assert msg

            print("\t- {}".format(msg.get("id")))
            await self.delete_msg(
                msg.get("channel_id"),
                msg.get("id")
            )

    async def delete_from_all(self):
        """

        Delete messages from the user's entire history by using a data download
        package.

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