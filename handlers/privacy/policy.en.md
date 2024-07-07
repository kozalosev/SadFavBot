The most authoritative, actual and comprehensive source of the truth, what information the bot is collected and stored, is [database migrations](https://github.com/kozalosev/SadFavBot/tree/main/db/migrations), published, as the rest of the source code, on GitHub.

However, let's take a closer look at the most important parts shortly:
1️⃣ any personal data isn't stored by the bot in any kind — only Telegram ID and settings (language and search options);
2️⃣ the bot doesn't store any media content — only IDs to it on the servers of Telegram;
3️⃣ but texts and coordinates of locations are really stored into the bot's database and linked to users by their ID;
4️⃣ texts and locations, deleted by users, currently is still present in the database, but the links are removed (this is a known [bug](https://github.com/kozalosev/SadFavBot/issues/64) planned to be fixed in the future);
5️⃣ obviously, all aliases and package names are also stored in the database;
6️⃣ none of the data, described above, is provided to third party organizations, except for:
   ➖ Russian hosting company [TimeWeb](https://timeweb.cloud), providing cloud infrastructure to run the code of applications;
   ➖ French hosting company [Scaleway](https://www.scaleway.com/en/), providing S3 compatible storage to store backups of the database;
   ➖ international company [Grafana Labs](https://grafana.com), in whose cloud infrastructure metrics and logs are stored.