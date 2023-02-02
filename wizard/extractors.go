package wizard

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
)

type FieldExtractor func(msg *tgbotapi.Message) interface{}

type File struct {
	FileID       string
	FileUniqueID string
}

func nilExtractor(*tgbotapi.Message) interface{}    { return nil }
func textExtractor(m *tgbotapi.Message) interface{} { return m.Text }
func stickerExtractor(m *tgbotapi.Message) interface{} {
	return File{FileID: m.Sticker.FileID, FileUniqueID: m.Sticker.FileUniqueID}
}
func voiceExtractor(m *tgbotapi.Message) interface{} {
	return File{FileID: m.Voice.FileID, FileUniqueID: m.Voice.FileUniqueID}
}
func audioExtractor(m *tgbotapi.Message) interface{} {
	return File{FileID: m.Audio.FileID, FileUniqueID: m.Audio.FileUniqueID}
}
func videoExtractor(m *tgbotapi.Message) interface{} {
	return File{FileID: m.Video.FileID, FileUniqueID: m.Video.FileUniqueID}
}
func videoNoteExtractor(m *tgbotapi.Message) interface{} {
	return File{FileID: m.VideoNote.FileID, FileUniqueID: m.VideoNote.FileUniqueID}
}
func gifExtractor(m *tgbotapi.Message) interface{} {
	return File{FileID: m.Animation.FileID, FileUniqueID: m.Animation.FileUniqueID}
}
func imageExtractor(m *tgbotapi.Message) interface{} {
	photo := m.Photo[len(m.Photo)-1]
	return File{FileID: photo.FileID, FileUniqueID: photo.FileUniqueID}
}

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
		log.Warningf("No action was found for %+v", msg)
		f.extractor = nilExtractor
	}
}
