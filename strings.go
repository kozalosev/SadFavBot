package main

func init() {
	locpool.Resources["en"] = map[string]string{
		"commands.save.fields.name":    "enter a name for your fav",
		"commands.save.fields.object":  "send an object to me",
		"commands.save.status.success": "saved successfully",
		"wizard.errors.state.missing":  "the state of your operation is missing; probably, the bot was restarted; please, try again from the beginning",
		"inline.errors.type.invalid":   "unknown type; contact me if you want it to be supported: @kozalo",
	}

	locpool.Resources["ru"] = map[string]string{
		"commands.save.fields.name":    "введите имя для закладки",
		"commands.save.fields.object":  "отправьте объект",
		"commands.save.status.success": "успешно сохранено",
		"try again from the beginning": "состояние операции потеряно; возможно, бот был перезапущен; пожалуйста, попробуйте повторить операцию с самого начала",
		"inline.errors.type.invalid":   "неизвестный тип; свяжитесь со мной, если хотите, чтобы он поддерживался: @kozalo",
	}
}
