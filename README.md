# `discord-delete`
A Python script capable of completely deleting Discord message history, even from channels you no longer participate in, using data request packages. It can also just clear messages from channels you are *currently* participating in. The script uses undocumented endpoints which are not permitted for developer use, so be warned that **using this script could result in the termination of your account** (see [self-bots](https://support.discordapp.com/hc/en-us/articles/115002192352-Automated-user-accounts-self-bots-)).

# Usage.
* Activate `virtualenv`
* Execute `pip install -r requirements.txt`
* Execute `python3 discord-delete.py -h` for a list of arguments

# Why?
I don't trust Discord with my personal data. They aren't profitable, and it's likely they will be acquired by another social media giant in future, such as Facebook. Discord anonymises accounts on deletion, but message history can usually be used to counteract that. They refuse to delete message history because it could make public conversations look confusing for other users. I really like Discord, but in my opinion, they don't take privacy seriously (why isn't there E2EE?). 