package main

func init() {
	locpool.Resources["en"] = map[string]string{
		"errors.unknown":                   "unexpected error has been occurred; please, try again later or contact @kozalo",
		"commands.save.fields.alias":       "enter a name for your fav",
		"commands.save.fields.object":      "send an object to me",
		"commands.save.status.success":     "saved successfully",
		"commands.save.status.failure":     "an error occurred while saving; please, try again later or contact @kozalo",
		"commands.delete.fields.alias":     "enter an alias",
		"commands.delete.fields.deleteAll": "Do you want to delete all objects associated with this alias?",
		"commands.delete.fields.object":    "send me an an object",
		"commands.delete.status.success":   "deleted successfully",
		"commands.delete.status.failure":   "an error occurred while deletion; please, try again later or contact @kozalo",
		"commands.delete.status.no.rows":   "such object wasn't found",
		"wizard.errors.state.missing":      "the state of your operation is missing; probably, the bot was restarted; please, try again from the beginning",
		"inline.errors.type.invalid":       "unknown type; contact me if you want it to be supported: @kozalo",
	}

	locpool.Resources["ru"] = map[string]string{
		"errors.unknown":                   "произошла неожиданная ошибка; пожалуйста, попробуйте позже или напишите @kozalo",
		"commands.save.fields.alias":       "введите имя для закладки",
		"commands.save.fields.object":      "отправьте объект",
		"commands.save.status.success":     "успешно сохранено",
		"commands.save.status.failure":     "при сохранении произошла ошибка, попробуйте позже или напишите @kozalo",
		"commands.delete.fields.alias":     "введите имя",
		"commands.delete.fields.deleteAll": "Вы хотите удалить все объекты под этим именем?",
		"commands.delete.fields.object":    "отправьте удаляемый объект",
		"commands.delete.status.success":   "успешно удалено",
		"commands.delete.status.failure":   "при удалении произошла ошибка, пожалуйста, попробуйте позже или напишите @kozalo",
		"commands.delete.status.no.rows":   "такой объект не был найден",
		"try again from the beginning":     "состояние операции потеряно; возможно, бот был перезапущен; пожалуйста, попробуйте повторить операцию с самого начала",
		"inline.errors.type.invalid":       "неизвестный тип; свяжитесь со мной, если хотите, чтобы он поддерживался: @kozalo",
	}
}
