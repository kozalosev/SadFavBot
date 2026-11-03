*Group and supergroup chats*

The bot may be added and used in group chats with some restrictions.

`/save` and `/link` ask for the missing data just like in a private chat: answer by replying to the question of the bot (your app opens the reply by itself). Use `/cancel` to stop. If the bot is an admin with the right to delete messages, its messages disappear from the chat after a while, together with your commands and answers; otherwise, everything stays.

The other commands work only in their short forms, i.e. one-liners like `/delete@SadFavBot alias` as a reply to some message. Called without the necessary data, they show a hint visible only to you. Note, if you call several commands in a row, at least the first one must contain the bot's name (unless the bot is promoted to an admin).

To save more objects under the same alias after `/save`, send them as replies to a message of the bot, e.g. to its question for the object, until the command expires.

Only a subset of commands is supported:
{{commands}}