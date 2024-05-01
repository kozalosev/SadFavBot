*Групповые чаты и супергруппы*

Бота можно добавлять и использовать в групповых чатах, но с некоторыми ограничениями. Наиболее важное из них, что команды должны вызываться только в коротких формах в виде одного сообщения: например, `/save алиас` как ответ на какое-либо сообщение.

Во-вторых, невозможно использовать цепочки объектов для сохранения под одним псевдонимом. Это очевидно из предыдущего ограничения.

В-третьих, поддерживается лишь часть команд:
{{commands}}

В текущей реализации бот игнорирует все некорректные формы команд, не содержащие всю необходимую информацию в момент вызова. Это может быть исправлено в будущем, чтобы упростить взаимодействие с ботом.