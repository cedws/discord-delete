#!/usr/bin/env python

import asyncio
import logging
import os

from io import StringIO
from csv import DictReader
from zipfile import ZipFile, BadZipFile
from argparse import ArgumentParser
from aiohttp import ClientSession

PARSER = ArgumentParser(
    description="Powerful script to delete full Discord message history"
)

PARSER.add_argument(
    "-v", "--verbose",
    action="store_true",
    help="enable verbose logging"
)

SUBCOMMAND = PARSER.add_subparsers(
    dest="cmd"
)

SUBCOMMAND.add_parser(
    "partial",
    help="run a partial message deletion."
)

SUBCOMMAND.add_parser(
    "full",
    help="run a full message deletion using a data request package"
).add_argument(
    "-p", "--package",
    required=True,
    help="path to the data request package."
)

API = "https://discordapp.com/api/v6"
LIMIT = 25
ENDPOINTS = {
    "me":               "/users/@me",
    "relationships":    "/users/@me/relationships",
    "guilds":           "/users/@me/guilds",
    "guild_msgs":       (
        "/guilds/{}/messages/search"
        "?author_id={}"
        "&include_nsfw=true"
        "&offset={}"
        "&limit={}"
    ),
    "channels":         "/users/@me/channels",
    "channel_msgs":     (
        "/channels/{}/messages/search"
        "?author_id={}"
        "&include_nsfw=true"
        "&offset={}"
        "&limit={}"
    ),
    "delete_msg":       "/channels/{}/messages/{}"
}

class Discord:
    def __init__(self, token):
        self.session = ClientSession(headers={
            "Authorization": token
        })

    async def __aenter__(self):
        return self

    async def __aexit__(self, *args):
        await self.session.close()

    async def _req(self, method, endpoint, **kwargs):
        logging.debug("%s %s", method, endpoint)

        url = "{}/{}".format(API, endpoint)

        async with self.session.request(method, url, **kwargs) as resp:
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
                delay = data["retry_after"]
                assert delay > 0

                logging.debug("Hit rate limit, waiting for a while.")
                await asyncio.sleep(delay / 1000)

                return await self._req(method, endpoint, **kwargs)
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
                logging.warning("Unknown status was sent.")
                logging.debug(await resp.text())

            return data

    async def me(self):
        return await self._req(
            "GET",
            ENDPOINTS["me"]
        )

    async def relationships(self):
        return await self._req(
            "GET",
            ENDPOINTS["relationships"]
        )

    async def relationship_channels(self, r_id):
        return await self._req(
            "POST",
            ENDPOINTS["channels"],
            json={"recipients": [r_id]}
        )

    async def guilds(self):
        return await self._req(
            "GET",
            ENDPOINTS["guilds"]
        )

    async def guild_msgs(self, g_id, a_id, offset=0):
        return await self._req(
            "GET",
            ENDPOINTS["guild_msgs"].format(g_id, a_id, offset, LIMIT)
        )

    async def channels(self):
        return await self._req(
            "GET",
            ENDPOINTS["channels"]
        )

    async def channel_msgs(self, c_id, a_id, offset=0):
        return await self._req(
            "GET",
            ENDPOINTS["channel_msgs"].format(c_id, a_id, offset, LIMIT)
        )

    async def delete_msg(self, c_id, m_id):
        return await self._req(
            "DELETE",
            ENDPOINTS["delete_msg"].format(c_id, m_id)
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
            rc_id = (await self.relationship_channels(r_id))["id"]

            await self.delete_from(self.channel_msgs, me_id, rc_id)

        # Gather a list of guilds (known by many as "servers"). This doesn't
        # include guilds that the user has left.
        guilds = await self.guilds()
        g_ids = [g.get("id") for g in guilds]

        for g_id in g_ids:
            await self.delete_from(self.guild_msgs, me_id, g_id)

    async def delete_from(self, get_msgs, me_id, u_id):
        logging.info("Deleting messages from %s...", u_id)

        messages = (await get_msgs(u_id, me_id)).get("messages")
        offset = 0

        while messages:
            # We avoid using the "total_results" field as it is often
            # inaccurate.
            logging.debug("Found %d more messages to process.", len(messages))

            for context in messages:
                # The "hit" field is the highlighted message. The other
                # messages are for context and may not be authored by the
                # user, so we ignore them. At least one message in the
                # context group should have the "hit" field.
                msg = next(m for m in context if m.get("hit"))
                assert msg

                if msg.get("type") != 0:
                    offset += 1
                    continue

                logging.info("- %s", msg.get("id"))
                await self.delete_msg(
                    msg.get("channel_id"),
                    msg.get("id")
                )

            messages = (await get_msgs(u_id, me_id, offset)).get("messages")

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

        me = await self.me()
        if not me:
            logging.error(
                "Failed to get user, auth token might have expired or "
                "be invalid."
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

            logging.info("Deleting messages from %s...", c_id)

            for line in reader:
                msg = line["ID"]

                logging.info("\t- %s", msg)
                await self.delete_msg(c_id, msg)

async def main():
    args = PARSER.parse_args()

    level = logging.DEBUG if args.verbose else logging.INFO
    logging.basicConfig(level=level)

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

if __name__ == "__main__":
    LOOP = asyncio.get_event_loop()
    LOOP.run_until_complete(main())
