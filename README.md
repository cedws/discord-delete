# discord-delete
A script capable of completely deleting Discord message history, even from channels you no longer participate in. Be warned that **using this script could result in the termination of your account** (see [self-bots](https://support.discordapp.com/hc/en-us/articles/115002192352-Automated-user-accounts-self-bots-)).

This project is vastly more efficient than others (which usually iterate through thousands of messages and hence take an extremely long time) since it intelligently uses undocumented endpoints to track down messages with precision. It keeps the number of API calls to the absolute minimum to reduce the risk of account termination. It's also able to do a deeper search for messages than other projects by using data request packages to delete messages from long-forgotten conversations.

# Usage
* (optional) Activate `virtualenv`
* Execute `pip3 install -r requirements.txt`
* Find your Discord authentication token using your browser's developer tools
* Execute `DISCORD_TOKEN=... python3 discord-delete.py` substituting `...` with the token to begin deletion

# Why?
I don't trust Discord with my personal data. They aren't profitable and therefore it's likely they will be acquired by another social media giant in future such as Facebook. Discord does anonymise accounts on deletion but message history can usually be used to counteract that. They refuse to delete message history with the justification that it could make public conversations look confusing for other users. From this, it's quite clear that Discord unfortunately does not take digital privacy seriously.