# discord-delete
A tool for deleting Discord message history. Be warned that **using this tool could result in the termination of your account**, though I've never had this happen even after extensive usage (see [self-bots](https://support.discordapp.com/hc/en-us/articles/115002192352-Automated-user-accounts-self-bots-)). The tool tries to interact with Discord's API efficiently to reduce the chance of this happening.

# Features
- Fast and efficient deletions
- Automatic token retrieval (currently for Windows only)

# Usage
- [Running a partial deletion](https://github.com/adversarialtools/discord-delete/wiki/Running-a-partial-deletion)

# Why?
Discord does not take a strong stance on privacy, unlike many other IM platforms that exist today, such as [Riot](https://about.riot.im/). This is visible from the choices they've made in designing their platform:
- No end-to-end encryption
- No built-in ability to delete full message history (not even on account deletion - they will only "anonymise" the messages)
- Demanding a mobile phone number from the user, in some cases in order to continue using the account
- Outright banning the use of alternative clients (which perhaps omit the telemetry that the proprietary client has)
- Encouraging users to "connect" other profiles, presumably to gain more data about the user

...amongst other egregious infringements on privacy, which you can find a list of [here](https://privacyspy.org/product/discord/).