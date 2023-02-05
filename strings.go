package main

func init() {
	locpool.Resources["en"] = map[string]string{
		"error":                       "Error",
		"commands.default.message":    "No command was selected. Send /help to know about my skills",
		"commands.save.fields.alias":  "enter a name for your fav",
		"commands.save.fields.object": "send an object to me",
		"commands.save.fields.alias.validation.error.length":  "the maximum length of the name must be less than %d characters",
		"commands.save.fields.object.validation.error.length": "the maximum length of the text must be less than %d characters",
		"commands.save.status.success":                        "saved successfully",
		"commands.save.status.failure":                        "an error occurred while saving; please, try again later or contact @kozalo",
		"commands.save.status.duplicate":                      "this item is already present in the collection",
		"commands.delete.fields.alias":                        "enter an alias",
		"commands.delete.fields.deleteAll":                    "Do you want to delete all objects associated with this alias?",
		"commands.delete.fields.object":                       "send me an an object",
		"commands.delete.status.success":                      "deleted successfully",
		"commands.delete.status.failure":                      "an error occurred while deletion; please, try again later or contact @kozalo",
		"commands.delete.status.no.rows":                      "such object wasn't found",
		"wizard.errors.field.invalid.value":                   "validation error: ",
		"wizard.errors.field.invalid.type":                    "the following type was expected: ",
		"wizard.errors.state.missing":                         "the state of your operation is missing; probably, the bot was restarted; please, try again from the beginning",
		"inline.errors.type.invalid":                          "unknown type; contact me if you want it to be supported: @kozalo",

		"video_note": "video note",
	}

	locpool.Resources["ru"] = map[string]string{
		"error":                       "Ошибка",
		"commands.default.message":    "Ни одна команда не была выбрана. Отправьте /help, чтобы узнать о моих возможностях",
		"commands.save.fields.alias":  "введите имя для закладки",
		"commands.save.fields.object": "отправьте объект",
		"commands.save.fields.alias.validation.error.length":  "имя не может быть длиннее %d символов",
		"commands.save.fields.object.validation.error.length": "текст не может быть длиннее %d символов",
		"commands.save.status.success":                        "успешно сохранено",
		"commands.save.status.failure":                        "при сохранении произошла ошибка, попробуйте позже или напишите @kozalo",
		"commands.save.status.duplicate":                      "данный объект уже присутствует в коллекции",
		"commands.delete.fields.alias":                        "введите имя",
		"commands.delete.fields.deleteAll":                    "Вы хотите удалить все объекты под этим именем?",
		"commands.delete.fields.object":                       "отправьте удаляемый объект",
		"commands.delete.status.success":                      "успешно удалено",
		"commands.delete.status.failure":                      "при удалении произошла ошибка, пожалуйста, попробуйте позже или напишите @kozalo",
		"commands.delete.status.no.rows":                      "такой объект не был найден",
		"wizard.errors.field.invalid.value":                   "ошибка валидации: ",
		"wizard.errors.field.invalid.type":                    "ожидался следующий тип: ",
		"wizard.errors.state.missing":                         "состояние операции потеряно; возможно, бот был перезапущен; пожалуйста, попробуйте повторить операцию с самого начала",
		"inline.errors.type.invalid":                          "неизвестный тип; свяжитесь со мной, если хотите, чтобы он поддерживался: @kozalo",

		"text":       "текстовое сообщение",
		"image":      "изображение",
		"gif":        "гифка",
		"video":      "видеозапись",
		"video_note": "видеосообщение",
		"sticker":    "стикер",
		"voice":      "голосовое сообщение",
		"audio":      "аудиозапись",
	}
}
