# discord-delete
A tool for deleting Discord message history.

>:warning: **Any tool for deleting messages, including this one, could result in the termination of your account** (see [self-bots](https://support.discordapp.com/hc/en-us/articles/115002192352-Automated-user-accounts-self-bots-)).
>Discord appears to have recently started cracking down on self-bots.

## Features
- Fast and efficient deletions
- Automatic token retrieval from Discord client (Windows, macOS, Linux)
- Dry run mode

## Usage
- [Running a partial deletion](https://github.com/adversarialtools/discord-delete/wiki/Running-a-partial-deletion)
- [Skipping specific channels](https://github.com/adversarialtools/discord-delete/wiki/Skipping-specific-channels)

## Why?
Discord does not take a strong stance on privacy, unlike many other IM platforms that exist today, such as [Matrix](https://matrix.org/). This is visible from the choices they've made in designing their platform:
- No end-to-end encryption
- No built-in ability to delete full message history (not even on account deletion - they will only "anonymise" the messages)
- Demanding a mobile phone number from the user, in some cases in order to continue using the account
- Outright banning the use of alternative clients (which perhaps omit the telemetry that the proprietary client has)
- Encouraging users to "connect" other profiles, presumably to gain more data about the user

...amongst other egregious infringements on privacy, which you can find a list of [here](https://privacyspy.org/product/discord/) and [here](https://cadence.moe/blog/2020-06-06-why-you-shouldnt-trust-discord).

### Note
It is recommended to first open all your DMs with tools such as [discord-dms-opener](https://github.com/Empyreann/discord-dms-opener) and then proceed for deletion.
