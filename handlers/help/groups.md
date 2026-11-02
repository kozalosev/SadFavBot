*Group and supergroup chats*

The bot may be added and used in group chats with some restrictions.

`/save` and `/link` ask for the missing data just like in a private chat: answer by replying to the question of the bot (your app opens the reply by itself). Use `/cancel` to stop. If the bot is an admin with the right to delete messages, the questions and your answers disappear from the chat after a while; otherwise, they stay. A question you haven't answered disappears in any case.

The other commands work only in their short forms, i.e. one-liners like `/delete@SadFavBot alias` as a reply to some message. Called without the necessary data, they show a hint visible only to you. Note, if you call several commands in a row, at least the first one must contain the bot's name (unless the bot is promoted to an admin).

You cannot use a chain of objects to be saved with the same alias: one command saves one object.

Only a subset of commands is supported:
{{commands}}