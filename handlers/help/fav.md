*Fav*

Fav is a favorite, i.e. an object or objects associated with some alias. The object may be of any type supported by Telegram messages:
— stickers
— images
— gifs
— voices
— music
— video
— video messages (notes)
— documents
— and even plain text (up to 4096 characters long)

In fact, the bot is a "key-value" (KV) storage. To create a fav, use the /save command, or /delete one for deletion. To print a list of your favs, use the /list command. You can filter the result by using it like this: `/list favs [part of alias]` (or `/list f [alias]` for short).

Also, there is a /ref command. Use it to know the aliases and packages associated with some object you sent (or you can just reply with this command to a message containing the object). If you want to know the packages containing some alias, use the short form of the command: `/ref [alias]`.