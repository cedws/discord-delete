import aiohttp
import asyncio
import logging
import os

from argparse import ArgumentParser
from zipfile import ZipFile, BadZipFile
from io import StringIO
from csv import DictReader

parser = ArgumentParser(
    description="Powerful script to delete full Discord message history"
)

parser.add_argument(
    "-v", "--verbose",
    action="store_true",
    help="enable verbose logging"
)

subcommand = parser.add_subparsers(
    dest="cmd"
)

partial = subcommand.add_parser(
    "partial",
    help="run a partial message deletion."
)

full = subcommand.add_parser(
    "full",
    help="run a full message deletion using a data request package"
)

full.add_argument(
    "-p", "--package",
    required=True,
    help="path to the data request package."
)

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
        logging.debug("%s %s", method, endpoint)

        if cache:
            cached = self.cache.get(method, endpoint, body=body)

            if cached:
                # We've made a request like this already and it's safe to return
                # the cached response.
                logging.debug("Using cached response.")
                return cached.json

        url = "{}/{}".format(API, endpoint)
        headers = { "Authorization": self.token }

        async with self.session.request(method, url, headers=headers,
                json=body) as resp:

            logging.debug("Got status %d from server.", resp.status)

            data = {}

            # Internal Server Error
            if resp.status >= 500:
                logging.error("Server reported internal error, aborting.")
                exit()
            # Too Many Requests
            elif resp.status == 429:
                data = await resp.json()

                # If we're being rate limited, wait for a while.
                delay = data.get("retry_after")
                assert delay > 0

                logging.debug("Hit rate limit, waiting for a while.")
                await asyncio.sleep(delay / 1000)

                return await self._req(method, endpoint, body)
            # Not Found
            elif resp.status == 404:
                logging.warning(
                    "Server sent Not Found. Resource may have been "
                    "deleted."
                )
            # Forbidden
            elif resp.status == 403:
                logging.warning("Server sent Forbidden.")
            # Unauthorized
            elif resp.status == 401:
                logging.warning("Server sent Unauthorized.")
            # No Content
            elif resp.status == 204:
                pass
            # OK
            elif resp.status == 200:
                data = await resp.json()
            else:
                logging.debug("Unknown status, response will not be cached.")
                return data

            if cache:
                logging.debug("Caching response.")
                self.cache.put(method, endpoint, data, body=body)

            return data

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
        if not me_id:
            logging.error(
                "Failed to get user, auth token might have expired or "
                "be invalid."
            )
            return

        logging.debug("User ID is %s.", me_id)

        # Gather a list of channels that the user is in. Channels are direct
        # messages which may be from groups or 1-to-1. Direct message
        # channels that have been hidden are not returned by the API.
        # We know the API keeps hidden channels open since it returns the
        # same channel ID if the channel is reopened.
        channels = await self.channels()
        c_ids = [c.get("id") for c in channels]

        for c_id in c_ids:
            await self.delete_from(self.channel_msgs, me_id, c_id)

        # Gather a lot of relationships the user has. These are people that the
        # user has in one of their lists (All/Pending/Blocked).
        relations = await self.relationships()
        r_ids = [r.get("id") for r in relations]

        # This code gathers a list of recipients who the user is in contact with,
        # but includes only those who are the *sole* recipient (that the
        # user might be friends with).
        c_recipients = [
            c.get("recipients")[0]
            for c in channels
            if len(c.get("recipients")) == 1
        ]
        # Get the ID for each recipient from the list above.
        c_recipient_ids = [c_r.get("id") for c_r in c_recipients]

        for r_id in r_ids:
            # Avoid making unnecessary requests by seeing if we already
            # checked the recipient channel earlier. Checking relationship
            # channels requires double the requests so we need to avoid doing it
            # if possible.
            if r_id in c_recipient_ids:
                logging.debug("Skipped recipient channel %s.", r_id)
                continue

            # Get the direct message channel with this recipient.
            # This wouldn't necessarily have been found earlier because
            # hidden channels aren't returned by the API.
            rc_id = await self.relationship_channels(r_id).get("id")

            # There should be a channel ID in all cases.
            assert rc_id

            await self.delete_from(self.channel_msgs, me_id, rc_id)

        # Gather a list of guilds (known by many as "servers"). This doesn't
        # include guilds that the user has left.
        guilds = await self.guilds()
        g_ids = [g.get("id") for g in guilds]

        for g_id in g_ids:
            await self.delete_from(self.guild_msgs, me_id, g_id)

    async def delete_from(self, get_msgs, me_id, u_id):
        logging.info("Deleting messages from %s...", u_id)

        # Avoids making unnecessary requests if the number of messages returned
        # is less that LIMIT, indicating there isn't another page of messages.
        results = LIMIT

        # Note the -1, there seems to be some kind of bug in the API where it
        # will sometimes return one less than the results limit.
        # TODO: Look into a better workaround.
        while results >= LIMIT - 1:
            mlist = await get_msgs(u_id, me_id)
            messages = mlist.get("messages")

            if not messages:
                break

            # We avoid using the "total_results" field as it is often
            # inaccurate.
            results = len(messages)
            logging.debug("Found %d more messages to process.", results)

            for context in messages:
                # The "hit" field is the highlighted message. The other
                # messages are for context and may not be authored by the
                # user, so we ignore them. At least one message in the
                # context group should have the "hit" field.
                msg = next(m for m in context if m.get("hit"))
                assert msg

                logging.info("\t- %s", msg.get("id"))
                await self.delete_msg(
                    msg.get("channel_id"),
                    msg.get("id")
                )

    async def delete_from_all(self, path):
        """

        Delete messages from the user's entire history by using a data request
        package.

        """

        if not os.path.exists(path):
            logging.error(
                "The specified data request package does not "
                "exist."
            )
            return

        try:
            data = ZipFile(path)
        except BadZipFile:
            logging.error(
                "The specified data request package is invalid "
                "(bad ZIP file)."
            )
            return

        channels = [
            file for file in data.namelist()
            if "messages.csv" in file
        ]

        logging.debug("Found %d channels to search.", len(channels))

        for channel in channels:
            msgs = StringIO(data.read(channel).decode("utf8"))
            reader = DictReader(msgs)

            c_id = channel.split("/")[1]
            assert c_id

            logging.info("Deleting messages from %s...", c_id)

            for line in reader:
                msg = line["ID"]
                assert msg

                logging.info("\t- %s", msg)
                await self.delete_msg(c_id, msg)

async def main():
    args = parser.parse_args()

    if args.verbose:
        logging.basicConfig(level=logging.DEBUG)
    else:
        logging.basicConfig(level=logging.INFO)

    # TODO: More secure way of passing this value.
    token = os.environ.get("DISCORD_TOKEN")

    if not token:
        logging.info(
            "You must pass a Discord auth token by setting DISCORD_TOKEN."
        )
        return

    async with Discord(token) as client:
        if args.cmd == "partial":
            await client.delete_from_current()
        if args.cmd == "full":
            await client.delete_from_all(args.package)

loop = asyncio.get_event_loop()
loop.run_until_complete(main())