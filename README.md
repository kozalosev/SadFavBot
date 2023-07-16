[@SadFavBot][SadFavBot-tg] — SadBot's Favorites Bot
===================================================

[![CI Build](https://github.com/kozalosev/SadFavBot/actions/workflows/ci-build.yml/badge.svg?branch=main&event=push)](https://github.com/kozalosev/SadFavBot/actions/workflows/ci-build.yml)
[![Wiki Docs](https://img.shields.io/badge/wiki-documentation-brightgreen)](../../wiki)

Your lovely pocket for storing stickers, pictures, GIFs, voices, etc. It lets you save your favorite memes and keep
them for further use via inline mode.

Save your favs:

![image](https://user-images.githubusercontent.com/25857981/234774715-8aaa7762-2c7d-4068-aa89-794dd91f6637.png)

And use them later!

![image](https://user-images.githubusercontent.com/25857981/230475842-e5a457ca-f903-4e53-82a6-3ae91f88584d.png)

Commands
--------

* `/help` — prints help message
* `/save` — starts a wizard that helps you store a meme associated with some textual alias
* `/link` — starts a wizard that helps you link another phrase to the already existing alias
* `/list` — prints all saved aliases or packages
* `/delete` — starts a wizard that helps you delete some memes you no longer needed
* `/hide` — exclude an alias from all listings
* `/package` — starts a wizard that helps you create or delete your own package of aliases, which may be installed by other users
* `/install` — command to install a package of aliases created by someone else
* `/language` — switch the bot's language
* `/mode` — enable or disable substring search
* `/cancel` — aborts the current wizard

Inline mode
-----------

Just write one of your saved aliases and get the stored memes!

How packages work?
------------------

The package is a set of aliases associated with the owner of the package. The package contains only aliases, not objects
associated with them! Therefore, the list of installing objects is formed only at installation time. The package is a
reference to specific aliases of the user. However, the installation is just a bunch of `/save` actions (a bunch of
inserts to the database actually), so the installed aliases are completely independent and separated between users.

[SadFavBot-tg]: https://t.me/SadFavBot
