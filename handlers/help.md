Hello, *{{username}}*!

I'm your lovely pocket for storing stickers, pictures, GIFs, voices, etc. I'm supposed to let you save your favorite memes and keep them for further use via inline mode

*Commands*

/help — prints help message
/save — starts a wizard that helps you store a meme associated with some textual alias
/list — prints all saved aliases or packages
/delete — starts a wizard that helps you delete some memes you no longer needed
/package — starts a wizard that helps you create or delete your own package of aliases, which may be installed by other users
/install — command to install a package of aliases created by someone else
/language — switch the bot's language
/cancel — aborts the current wizard

*Inline mode*

Just write one of your saved aliases and get the stored memes!

*How packages work?*

The package is a set of aliases associated with the owner of the package. The package contains only aliases, not objects associated with them! Therefore, the list of installing objects is formed only at installation time. The package is a reference to specific aliases of the user. However, the installation is just a bunch of `/save` actions (a bunch of inserts to the database actually), so the installed aliases are completely independent and separated between users.
