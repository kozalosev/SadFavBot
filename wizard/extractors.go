package wizard

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type FieldExtractor func(msg *tgbotapi.Message) string

func textExtractor(m *tgbotapi.Message) string      { return m.Text }
func stickerExtractor(m *tgbotapi.Message) string   { return m.Sticker.FileID }
func imageExtractor(m *tgbotapi.Message) string     { return m.Photo[0].FileID }
func voiceExtractor(m *tgbotapi.Message) string     { return m.Voice.FileID }
func audioExtractor(m *tgbotapi.Message) string     { return m.Audio.FileID }
func videoExtractor(m *tgbotapi.Message) string     { return m.Video.FileID }
func videoNoteExtractor(m *tgbotapi.Message) string { return m.VideoNote.FileID }
func gifExtractor(m *tgbotapi.Message) string       { return m.Animation.FileID }

func determineMessageType(msg *tgbotapi.Message) FieldType {
	if msg.Sticker != nil {
		return Sticker
	}
	if msg.Photo != nil {
		return Image
	}
	if msg.Voice != nil {
		return Voice
	}
	if msg.Audio != nil {
		return Audio
	}
	if msg.Video != nil {
		return Video
	}
	if msg.VideoNote != nil {
		return VideoNote
	}
	if msg.Animation != nil {
		return Gif
	}
	return Text
}

func (f *Field) restoreExtractor(msg *tgbotapi.Message) {
	if f.extractor != nil {
		return
	}
	switch f.Type {
	case Auto:
		msgType := determineMessageType(msg)
		f.Type = msgType
		f.restoreExtractor(msg)
	case Text:
		f.extractor = textExtractor
	case Sticker:
		f.extractor = stickerExtractor
	case Image:
		f.extractor = imageExtractor
	case Voice:
		f.extractor = voiceExtractor
	case Audio:
		f.extractor = audioExtractor
	case Video:
		f.extractor = videoExtractor
	case VideoNote:
		f.extractor = videoNoteExtractor
	case Gif:
		f.extractor = gifExtractor
	default:
		f.extractor = func(m *tgbotapi.Message) string { return "no action was found!" } // elaborate more elegant solution
	}
}
