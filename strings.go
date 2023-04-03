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
		"error":                               "Error",
		"success":                             "👍👌",
		"commands.default.message":            "No command was selected. Send /help to know about my skills",
		"commands.default.message.on.command": "Unknown command. Send /help to know about my skills",
		"commands.start.status.failure":       "Something went wrong... Please, try again later or contact @kozalo",
		"commands.save.fields.alias":          "Enter a name for your fav",
		"commands.save.fields.object":         "Send an object to me",
		"commands.save.fields.alias.validation.error.length":              "the maximum length of the name must be less than %d characters",
		"commands.save.fields.alias.validation.error.forbidden.symbols":   trValidationForbiddenSymbolsEn,
		"commands.save.fields.object.validation.error.length":             "the maximum length of the text must be less than %d characters",
		"commands.save.status.success":                                    "Saved successfully",
		"commands.save.status.failure":                                    fmt.Sprintf(trStatusFailureTemplateEn, "saving"),
		"commands.save.status.duplicate":                                  "This item is already present in the collection",
		"commands.list.fields.aliases.or.packages":                        "Aliases or packages?",
		"commands.list.status.success.aliases":                            "Saved aliases (with the count of associated objects):",
		"commands.list.status.success.packages":                           "Created packages (with the count of associated aliases):",
		"commands.list.status.failure":                                    fmt.Sprintf(trStatusFailureTemplateEn, "executing the query"),
		"commands.list.status.no.rows.aliases":                            "You have no saved aliases. It's time to make use of the /save command!",
		"commands.list.status.no.rows.packages":                           "You have no packages. It's time to create a new one using the /package command!",
		"commands.delete.fields.alias":                                    "Enter an alias",
		"commands.delete.fields.deleteAll":                                "Do you want to delete all objects associated with this alias?",
		"commands.delete.fields.object":                                   "Send me an an object",
		"commands.delete.button.select.object":                            "Select…",
		"commands.delete.status.success":                                  "Deleted successfully",
		"commands.delete.status.failure":                                  fmt.Sprintf(trStatusFailureTemplateEn, "deletion"),
		"commands.delete.status.no.rows":                                  "Such object wasn't found",
		"commands.language.fields.language":                               "Choose your language:",
		"commands.language.status.failure":                                fmt.Sprintf(trStatusFailureTemplateEn, "saving your settings"),
		"commands.package.fields.createOrDelete":                          "Do you want to create or delete a package?",
		"commands.package.fields.name":                                    "Enter a name of your package",
		"commands.package.fields.aliases":                                 "Please, write a list of aliases (one per line) you want to include in the package. You can use the list generated by the /list command",
		"commands.package.fields.name.validation.error.length":            "the maximum length of the name must be less than %d characters",
		"commands.package.fields.name.validation.error.forbidden.symbols": trValidationForbiddenSymbolsEn,
		"commands.package.status.success.creation":                        "The package was created successfully! Its name is *%s*\nThe command to install: `/install %s`\nLink: https://t.me/%s?start=%s",
		"commands.package.status.success.deletion":                        "The package was deleted successfully",
		"commands.package.status.failure":                                 fmt.Sprintf(trStatusFailureTemplateEn, "creating the package"),
		"commands.package.status.duplicate":                               "A package with the same name already exists",
		"commands.package.status.no.rows":                                 "Nothing to delete",
		"commands.install.fields.name":                                    "Enter the name of a package",
		"commands.install.fields.confirmation":                            "Are you sure you want to install it?",
		"commands.install.status.success":                                 "The following aliases were installed successfully:",
		"commands.install.status.success.no.names":                        "Installed successfully",
		"commands.install.status.failure":                                 fmt.Sprintf(trStatusFailureTemplateEn, "installing the package"),
		"commands.install.status.no.rows":                                 "You already have all items from this package",
		"commands.install.message.package.items.count":                    "The count of aliases in the package %s is %d",
		"commands.link.fields.name":                                       "Enter a name for a link",
		"commands.link.fields.alias":                                      "Enter the name you wanna link to",
		"commands.link.status.success":                                    "The link was created successfully",
		"commands.link.status.failure":                                    fmt.Sprintf(trStatusFailureTemplateEn, "creating the link"),
		"commands.link.status.duplicate":                                  "A link with the same name already exists",
		"wizard.active.not.set":                                           "There is nothing to cancel ☹",
		"wizard.errors.field.invalid.value":                               "Validation error: ",
		"wizard.errors.field.invalid.type":                                "The following type was expected: ",
		"wizard.errors.state.missing":                                     "The state of your operation is missing; probably, the bot was restarted; please, try again from the beginning",
		"inline.errors.type.invalid":                                      "Unknown type; contact me if you want it to be supported: @kozalo",
		"errors.validation.option.not.in.list":                            "option is not from the suggested variants",
		"callbacks.error":                                                 "An unknown error has been occurred, try again later or contact @kozalo",

		"video_note": "video note",
	}

	locpool.Resources["ru"] = map[string]string{
		"error":                               "Ошибка",
		"success":                             "👍👌",
		"commands.default.message":            "Ни одна команда не была выбрана. Отправьте /help, чтобы узнать о моих возможностях",
		"commands.default.message.on.command": "Неизвестная команда. Отправьте /help, чтобы узнать о моих возможностях",
		"commands.start.status.failure":       "Что-то пошло не так... Пожалуйста, повторите позднее или напишите @kozalo",
		"commands.save.fields.alias":          "Введите имя для закладки",
		"commands.save.fields.object":         "Отправьте объект",
		"commands.save.fields.alias.validation.error.length":              "имя не может быть длиннее %d символов",
		"commands.save.fields.alias.validation.error.forbidden.symbols":   trValidationForbiddenSymbolsRu,
		"commands.save.fields.object.validation.error.length":             "текст не может быть длиннее %d символов",
		"commands.save.status.success":                                    "Успешно сохранено",
		"commands.save.status.failure":                                    fmt.Sprintf(trStatusFailureTemplateRu, "сохранении"),
		"commands.save.status.duplicate":                                  "Данный объект уже присутствует в коллекции",
		"commands.list.fields.aliases.or.packages":                        "Закладки или наборы?",
		"commands.list.status.success.aliases":                            "Сохранённые закладки (с количеством ассоциированных объектов):",
		"commands.list.status.success.packages":                           "Созданные наборы (с количеством ассоциированных закладок):",
		"commands.list.status.failure":                                    fmt.Sprintf(trStatusFailureTemplateRu, "выполнении запроса"),
		"commands.list.status.no.rows.aliases":                            "У вас нет никаких закладок. Самое время воспользоваться командой /save!",
		"commands.list.status.no.rows.packages":                           "У вас нет никаких наборов. Самое время создать новый с помощью команды /package!",
		"commands.delete.fields.alias":                                    "Введите имя",
		"commands.delete.fields.deleteAll":                                "Вы хотите удалить все объекты под этим именем?",
		"commands.delete.fields.object":                                   "Отправьте удаляемый объект",
		"commands.delete.button.select.object":                            "Выбрать…",
		"commands.delete.status.success":                                  "Успешно удалено",
		"commands.delete.status.failure":                                  fmt.Sprintf(trStatusFailureTemplateRu, "удалении"),
		"commands.delete.status.no.rows":                                  "Такой объект не был найден",
		"commands.language.fields.language":                               "Выберите свой язык:",
		"commands.language.status.failure":                                fmt.Sprintf(trStatusFailureTemplateRu, "сохранении выбранного языка"),
		"commands.package.fields.createOrDelete":                          "Вы хотите создать или удалить набор?",
		"commands.package.fields.name":                                    "Введите название для набора",
		"commands.package.fields.aliases":                                 "Пожалуйста, перечислите названия всех закладок (по одному на строчку), которые хотите включить в набор. Можно использовать список, сгенерированный командой /list",
		"commands.package.fields.name.validation.error.length":            "название не может быть длиннее %d символов",
		"commands.package.fields.name.validation.error.forbidden.symbols": trValidationForbiddenSymbolsRu,
		"commands.package.status.success.creation":                        "Набор успешно сохранён! Название: *%s*\nКоманда для установки: `/install %s`\nСсылка для установки: https://t.me/%s?start=%s",
		"commands.package.status.success.deletion":                        "Успешно удалено",
		"commands.package.status.failure":                                 fmt.Sprintf(trStatusFailureTemplateRu, "создании набора"),
		"commands.package.status.duplicate":                               "Набор с таким названием уже существует",
		"commands.package.status.no.rows":                                 "Нечего удалять",
		"commands.install.fields.name":                                    "Введите название набора",
		"commands.install.fields.confirmation":                            "Вы уверены, что хотите установить набор?",
		"commands.install.status.success":                                 "Следующие закладки были успешно установлены:",
		"commands.install.status.success.no.names":                        "Набор был успешно установлен",
		"commands.install.status.failure":                                 fmt.Sprintf(trStatusFailureTemplateRu, "установке набора"),
		"commands.install.status.no.rows":                                 "У вас уже сохранены все закладки из этого набора",
		"commands.install.message.package.items.count":                    "Количество псевдонимов в наборе __%s__ равняется %d",
		"commands.link.fields.name":                                       "Введите название для ссылки",
		"commands.link.fields.alias":                                      "Введите название закладки, на которую она будет ссылаться",
		"commands.link.status.success":                                    "Ссылка успешно сохранена",
		"commands.link.status.failure":                                    fmt.Sprintf(trStatusFailureTemplateRu, "создании ссылки"),
		"commands.link.status.duplicate":                                  "Ссылка с таким названием уже существует",
		"wizard.active.not.set":                                           "Нечего отменять ☹",
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

		"Create":   "Создать",
		"Delete":   "Удалить",
		"Aliases":  "Закладки",
		"Packages": "Наборы",
	}
}
