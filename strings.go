package main

import "fmt"

const (
	trStatusFailureTemplateEn      = "An error occurred while %s; please, try again later or contact @kozalo"
	trStatusFailureTemplateRu      = "При %s произошла ошибка, попробуйте позже или напишите @kozalo"
	trValidationForbiddenSymbolsEn = "The following symbols are forbidden for use: `%s` (+ new line character)"
	trValidationForbiddenSymbolsRu = "Следующие символы не могут быть использованы для названия: `%s` (+ перенос строки)"
)

func init() {
	locpool.Resources["en"] = map[string]string{
		"error":                                              "Error",
		"success":                                            "👍👌",
		"commands.default.message":                           "No command was selected. Send /help to know about my skills",
		"commands.default.message.on.command":                "Unknown command. Send /help to know about my skills",
		"commands.start.status.failure":                      "Something went wrong... Please, try again later or contact @kozalo",
		"commands.help.description":                          "show help message",
		"commands.cancel.description":                        "abort the current operation",
		"commands.save.description":                          "associate one or more objects with some alias",
		"commands.save.fields.alias":                         "Enter an alias for your fav",
		"commands.save.fields.object":                        "Send me an object",
		"commands.save.fields.alias.validation.error.length": "the maximum length of the alias must be less than %d characters",
		"commands.save.fields.alias.validation.error.forbidden.symbols":   trValidationForbiddenSymbolsEn,
		"commands.save.fields.object.validation.error.length":             "the maximum length of the text must be less than %d characters",
		"commands.save.status.success":                                    "Your fav has been saved successfully!\nType `@%s %s` in any chat to use it",
		"commands.save.status.failure":                                    fmt.Sprintf(trStatusFailureTemplateEn, "saving"),
		"commands.save.status.duplicate":                                  "This object is already present in your favs",
		"commands.list.description":                                       "print the list of all saved aliases or packages",
		"commands.list.fields.favs.or.packages":                           "Favs or packages?",
		"commands.list.status.success.favs":                               "Saved favs (with the count of associated objects):",
		"commands.list.status.success.packages":                           "Created packages (with the count of associated aliases):",
		"commands.list.status.failure":                                    fmt.Sprintf(trStatusFailureTemplateEn, "executing the query"),
		"commands.list.status.no.rows.favs":                               "You have no saved favs. It's time to make use of the /save command!",
		"commands.list.status.no.rows.packages":                           "You have no packages. It's time to create a new one using the /package command!",
		"commands.delete.description":                                     "delete some of your favs",
		"commands.delete.fields.alias":                                    "Enter an alias",
		"commands.delete.fields.deleteAll":                                "Do you want to delete all objects associated with this alias?",
		"commands.delete.fields.object":                                   "Send me an object",
		"commands.delete.button.select.object":                            "Select…",
		"commands.delete.status.success":                                  "Deleted successfully",
		"commands.delete.status.failure":                                  fmt.Sprintf(trStatusFailureTemplateEn, "deletion"),
		"commands.delete.status.no.rows":                                  "Such object wasn't found",
		"commands.language.description":                                   "change the language",
		"commands.language.fields.language":                               "Choose your language:",
		"commands.language.status.failure":                                fmt.Sprintf(trStatusFailureTemplateEn, "saving your settings"),
		"commands.package.description":                                    "create or delete a shareable package of your favs",
		"commands.package.fields.createOrDelete":                          "Do you want to create or delete a package?",
		"commands.package.fields.name":                                    "Enter package name",
		"commands.package.fields.aliases":                                 "Please, write a list of aliases (one per line) which you want to include in the package. You can use the list generated by the /list command",
		"commands.package.fields.name.validation.error.length":            "the maximum length of the name must be less than %d characters",
		"commands.package.fields.name.validation.error.forbidden.symbols": trValidationForbiddenSymbolsEn,
		"commands.package.status.success.creation":                        "The package has been created successfully! Its name is *%s*\nThe command to install: `/install %s`\nLink: https://t.me/%s?start=%s",
		"commands.package.status.success.recreation":                      "The package has been recreated successfully! Its name is *%s*\nThe command to install: `/install %s`\nLink: https://t.me/%s?start=%s",
		"commands.package.status.success.deletion":                        "The package has been deleted successfully",
		"commands.package.status.failure":                                 fmt.Sprintf(trStatusFailureTemplateEn, "creating the package"),
		"commands.package.status.duplicate":                               "A package with the same name already exists",
		"commands.package.status.no.rows":                                 "Nothing to delete",
		"commands.install.description":                                    "install someone else's package of his/her favs",
		"commands.install.fields.name":                                    "Enter the name of a package",
		"commands.install.fields.confirmation":                            "Are you sure you want to install it?",
		"commands.install.status.success":                                 "The following aliases were installed successfully:",
		"commands.install.status.success.no.names":                        "Installed successfully",
		"commands.install.status.failure":                                 fmt.Sprintf(trStatusFailureTemplateEn, "installing the package"),
		"commands.install.status.no.rows":                                 "You already have all favs from this package",
		"commands.install.status.link.existing.fav":                       "You're trying to add a link to the alias you already have as a fav. Installation of the packages containing links has some restrictions for now",
		"commands.install.message.package.favs":                           "The package _%s_ consists of:\n\n%s",
		"commands.link.description":                                       "create a link to the alias already present within your favs",
		"commands.link.fields.name":                                       "Enter a name for a link",
		"commands.link.fields.name.validation.error.length":               "the maximum length of the link name must be less than %d characters",
		"commands.link.fields.name.validation.error.forbidden.symbols":    trValidationForbiddenSymbolsEn,
		"commands.link.fields.alias":                                      "Enter the name you wanna link to",
		"commands.link.status.success":                                    "The link has been created successfully",
		"commands.link.status.failure":                                    fmt.Sprintf(trStatusFailureTemplateEn, "creating the link"),
		"commands.link.status.duplicate":                                  "A link with the same name already exists",
		"commands.link.status.duplicate.fav":                              "A fav with the same name already exists",
		"commands.link.status.no.alias":                                   "You're trying to link a non-existing fav",
		"commands.mode.description":                                       "enable or disable substring search",
		"commands.mode.fields.substringSearchEnabled":                     "Do you want to search by substring, not by exact match?",
		"commands.mode.status.success":                                    "Switched successfully",
		"commands.mode.status.failure":                                    fmt.Sprintf(trStatusFailureTemplateEn, "saving the value"),
		"commands.mode.message.current.value":                             "The current value of the substring search mode is ",
		"commands.visibility.description":                                 "exclude an alias from all listings, or reveal it back",
		"commands.visibility.fields.change":                               "Would you like to exclude an alias from all listings or reveal it?",
		"commands.visibility.fields.alias":                                "Enter an alias",
		"commands.visibility.status.failure":                              fmt.Sprintf(trStatusFailureTemplateEn, "changing the visibility of the alias"),
		"commands.visibility.status.no.rows":                              "The alias doesn't exist",
		"commands.ref.description":                                        "print the list of all aliases associated with an object",
		"commands.ref.fields.object":                                      "Send me an object",
		"commands.ref.status.success":                                     "Associated aliases:",
		"commands.ref.status.failure":                                     fmt.Sprintf(trStatusFailureTemplateRu, "executing the query"),
		"commands.ref.status.no.rows":                                     "There is no favs associated with this object. It's time to make use of the /save command!",
		"callbacks.help.button.fav":                                       "Fav",
		"callbacks.help.button.alias":                                     "Alias",
		"callbacks.help.button.inline":                                    "Search",
		"callbacks.help.button.package":                                   "Package",
		"callbacks.help.button.link":                                      "Link",
		"callbacks.help.button.settings":                                  "Settings",
		"callbacks.help.caption.inline":                                   "In other words, type the name of the bot after the `@` sign and your query afterward",
		"callbacks.help.message.current.page":                             "You're already on this page",
		"wizard.active.not.set":                                           "There is nothing to cancel 🙁",
		"wizard.errors.field.invalid.value":                               "Validation error: ",
		"wizard.errors.field.invalid.type":                                "The following type was expected: ",
		"wizard.errors.state.missing":                                     "The state of your operation is missing; probably, the bot was restarted; please, try again from the beginning",
		"inline.errors.type.invalid":                                      "Unknown type; contact me if you want it to be supported: @kozalo",
		"errors.validation.option.not.in.list":                            "option is not from the suggested variants",
		"callbacks.error":                                                 "An unknown error has been occurred, try again later or contact @kozalo",

		"video_note": "video note",
	}

	locpool.Resources["ru"] = map[string]string{
		"error":                                              "Ошибка",
		"success":                                            "👍👌",
		"commands.default.message":                           "Ни одна команда не была выбрана. Отправьте /help, чтобы узнать о моих возможностях",
		"commands.default.message.on.command":                "Неизвестная команда. Отправьте /help, чтобы узнать о моих возможностях",
		"commands.start.status.failure":                      "Что-то пошло не так... Пожалуйста, повторите позднее или напишите @kozalo",
		"commands.help.description":                          "показать сообщение с помощью",
		"commands.cancel.description":                        "прервать текущую операцию",
		"commands.save.description":                          "сохранить один или несколько объектов под определённым алиасом",
		"commands.save.fields.alias":                         "Введите название для закладки",
		"commands.save.fields.object":                        "Отправьте объект",
		"commands.save.fields.alias.validation.error.length": "название не может быть длиннее %d символов",
		"commands.save.fields.alias.validation.error.forbidden.symbols":   trValidationForbiddenSymbolsRu,
		"commands.save.fields.object.validation.error.length":             "текст не может быть длиннее %d символов",
		"commands.save.status.success":                                    "Ваша закладка успешно сохранена! Чтобы использовать её, наберите `@%s %s` в любом чате",
		"commands.save.status.failure":                                    fmt.Sprintf(trStatusFailureTemplateRu, "сохранении"),
		"commands.save.status.duplicate":                                  "Данный объект уже присутствует в коллекции",
		"commands.list.description":                                       "вывести список либо сохранённых алиасов, либо созданных пакетов",
		"commands.list.fields.favs.or.packages":                           "Закладки или пакеты?",
		"commands.list.status.success.favs":                               "Сохранённые закладки (с количеством ассоциированных объектов):",
		"commands.list.status.success.packages":                           "Созданные пакеты (с количеством ассоциированных закладок):",
		"commands.list.status.failure":                                    fmt.Sprintf(trStatusFailureTemplateRu, "выполнении запроса"),
		"commands.list.status.no.rows.favs":                               "У вас нет никаких закладок. Самое время воспользоваться командой /save!",
		"commands.list.status.no.rows.packages":                           "У вас нет никаких пакетов. Самое время создать новый с помощью команды /package!",
		"commands.delete.description":                                     "удалить некоторые закладки из Вашей коллекции",
		"commands.delete.fields.alias":                                    "Введите алиас",
		"commands.delete.fields.deleteAll":                                "Вы хотите удалить все объекты под этим алиасом?",
		"commands.delete.fields.object":                                   "Отправьте удаляемый объект",
		"commands.delete.button.select.object":                            "Выбрать…",
		"commands.delete.status.success":                                  "Успешно удалено",
		"commands.delete.status.failure":                                  fmt.Sprintf(trStatusFailureTemplateRu, "удалении"),
		"commands.delete.status.no.rows":                                  "Такой объект не был найден",
		"commands.language.description":                                   "сменить язык",
		"commands.language.fields.language":                               "Выберите свой язык:",
		"commands.language.status.failure":                                fmt.Sprintf(trStatusFailureTemplateRu, "сохранении выбранного языка"),
		"commands.package.description":                                    "создать или удалить пакет закладок, которыми Вы хотите поделиться с другими",
		"commands.package.fields.createOrDelete":                          "Вы хотите создать или удалить пакет?",
		"commands.package.fields.name":                                    "Введите название",
		"commands.package.fields.aliases":                                 "Пожалуйста, перечислите названия всех закладок (по одному на строчку), которые хотите включить в пакет. Можно использовать список, сгенерированный командой /list",
		"commands.package.fields.name.validation.error.length":            "название не может быть длиннее %d символов",
		"commands.package.fields.name.validation.error.forbidden.symbols": trValidationForbiddenSymbolsRu,
		"commands.package.status.success.creation":                        "Пакет успешно сохранён! Название: *%s*\nКоманда для установки: `/install %s`\nСсылка для установки: https://t.me/%s?start=%s",
		"commands.package.status.success.recreation":                      "Пакет успешно пересоздан! Название: *%s*\nКоманда для установки: `/install %s`\nСсылка для установки: https://t.me/%s?start=%s",
		"commands.package.status.success.deletion":                        "Пакет был успешно удалён",
		"commands.package.status.failure":                                 fmt.Sprintf(trStatusFailureTemplateRu, "создании пакета"),
		"commands.package.status.duplicate":                               "Пакет с таким названием уже существует",
		"commands.package.status.no.rows":                                 "Нет такого пакета 🙁",
		"commands.install.description":                                    "установить закладки из пакета, созданного кем-то другим",
		"commands.install.fields.name":                                    "Введите название пакета",
		"commands.install.fields.confirmation":                            "Вы уверены, что хотите установить пакет?",
		"commands.install.status.success":                                 "Следующие закладки были успешно установлены:",
		"commands.install.status.success.no.names":                        "Пакет был успешно установлен",
		"commands.install.status.failure":                                 fmt.Sprintf(trStatusFailureTemplateRu, "установке пакета"),
		"commands.install.status.no.rows":                                 "У вас уже сохранены все закладки из этого пакета",
		"commands.install.status.link.existing.fav":                       "При установке этого пакета происходит попытка создать ссылку на алиас, который у Вас уже имеется в виде ссылки. Пока установка таких пакетов запрещена",
		"commands.install.message.package.favs":                           "Пакет _%s_ состоит из:\n\n%s",
		"commands.link.description":                                       "создать ссылку на алиас, уже имеющийся в Вашей коллекции",
		"commands.link.fields.name":                                       "Введите название для ссылки",
		"commands.link.fields.name.validation.error.length":               "название ссылки не может быть длиннее %d символов",
		"commands.link.fields.name.validation.error.forbidden.symbols":    trValidationForbiddenSymbolsRu,
		"commands.link.fields.alias":                                      "Введите название закладки, на которую она будет ссылаться",
		"commands.link.status.success":                                    "Ссылка успешно сохранена",
		"commands.link.status.failure":                                    fmt.Sprintf(trStatusFailureTemplateRu, "создании ссылки"),
		"commands.link.status.duplicate":                                  "Ссылка с таким названием уже существует",
		"commands.link.status.duplicate.fav":                              "Закладка с таким названием уже существует",
		"commands.link.status.no.alias":                                   "Вы пытаетесь создать ссылку для несуществующей закладки",
		"commands.mode.description":                                       "включить или выключить поиск по подстроке",
		"commands.mode.fields.substringSearchEnabled":                     "Включить режим поиска по подстроке, а не по полному совпадению?",
		"commands.mode.status.success":                                    "Значение параметра успешно изменено",
		"commands.mode.status.failure":                                    fmt.Sprintf(trStatusFailureTemplateRu, "сохранении параметра"),
		"commands.mode.message.current.value":                             "Текущее значение рубильника режима поиска по подстроке: ",
		"commands.visibility.description":                                 "скрыть алиас из отображения во всех списках или вернуть обратно",
		"commands.visibility.fields.change":                               "Вы хотите исключить алиас из всех списков или показывать его?",
		"commands.visibility.fields.alias":                                "Введите алиас",
		"commands.visibility.status.failure":                              fmt.Sprintf(trStatusFailureTemplateRu, "скрытии алиаса"),
		"commands.visibility.status.no.rows":                              "Алиаса не существует",
		"commands.ref.description":                                        "вывести список алиасов, связанных с объектом",
		"commands.ref.fields.object":                                      "Отправьте объект",
		"commands.ref.status.success":                                     "Связанные алиасы:",
		"commands.ref.status.failure":                                     fmt.Sprintf(trStatusFailureTemplateRu, "выполнении запроса"),
		"commands.ref.status.no.rows":                                     "Нет закладок для этого объекта. Самое время воспользоваться командой /save!",
		"callbacks.help.button.fav":                                       "Закладка",
		"callbacks.help.button.alias":                                     "Алиас",
		"callbacks.help.button.inline":                                    "Поиск",
		"callbacks.help.button.package":                                   "Пакет",
		"callbacks.help.button.link":                                      "Ссылка",
		"callbacks.help.button.settings":                                  "Настройки",
		"callbacks.help.caption.inline":                                   "Чтобы использовать inline-режим, нужно ввести название бота через собачку, а потом свой запрос",
		"callbacks.help.message.current.page":                             "Вы уже находитесь на этой странице",
		"wizard.active.not.set":                                           "Нечего отменять 🙁",
		"wizard.errors.field.invalid.value":                               "Ошибка валидации: ",
		"wizard.errors.field.invalid.type":                                "Ожидался следующий тип: ",
		"wizard.errors.state.missing":                                     "Состояние операции потеряно; возможно, бот был перезапущен; пожалуйста, попробуйте повторить операцию с самого начала",
		"inline.errors.type.invalid":                                      "Неизвестный тип; свяжитесь со мной, если хотите, чтобы он поддерживался: @kozalo",
		"errors.validation.option.not.in.list":                            "введённого варианта нет в предложенном списке",
		"callbacks.error":                                                 "Произошла неизвестная ошибка, попробуйте позже или напишите @kozalo",

		"text":       "текст",
		"image":      "изображение",
		"gif":        "гифка",
		"video":      "видеозапись",
		"video_note": "видеосообщение",
		"sticker":    "стикер",
		"voice":      "аудиосообщение",
		"audio":      "аудиозапись",
		"document":   "документ",
		"location":   "местоположение",

		"Create":   "Создать",
		"Recreate": "Пересоздать",
		"Delete":   "Удалить",
		"Favs":     "Закладки",
		"Packages": "Пакеты",

		"exclude": "исключить",
		"reveal":  "показывать",
	}
}
