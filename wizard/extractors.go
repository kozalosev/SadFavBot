package wizard

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
)

type FieldExtractor func(msg *tgbotapi.Message) interface{}

type File struct {
	ID       string
	UniqueID string
}

func nilExtractor(*tgbotapi.Message) interface{}    { return nil }
func textExtractor(m *tgbotapi.Message) interface{} { return m.Text }
func stickerExtractor(m *tgbotapi.Message) interface{} {
	if m.Sticker == nil {
		return nil
	}
	return File{ID: m.Sticker.FileID, UniqueID: m.Sticker.FileUniqueID}
}
func voiceExtractor(m *tgbotapi.Message) interface{} {
	if m.Voice == nil {
		return nil
	}
	return File{ID: m.Voice.FileID, UniqueID: m.Voice.FileUniqueID}
}
func audioExtractor(m *tgbotapi.Message) interface{} {
	if m.Audio == nil {
		return nil
	}
	return File{ID: m.Audio.FileID, UniqueID: m.Audio.FileUniqueID}
}
func videoExtractor(m *tgbotapi.Message) interface{} {
	if m.Video == nil {
		return nil
	}
	return File{ID: m.Video.FileID, UniqueID: m.Video.FileUniqueID}
}
func videoNoteExtractor(m *tgbotapi.Message) interface{} {
	if m.VideoNote == nil {
		return nil
	}
	return File{ID: m.VideoNote.FileID, UniqueID: m.VideoNote.FileUniqueID}
}
func gifExtractor(m *tgbotapi.Message) interface{} {
	if m.Animation == nil {
		return nil
	}
	return File{ID: m.Animation.FileID, UniqueID: m.Animation.FileUniqueID}
}
func imageExtractor(m *tgbotapi.Message) interface{} {
	if m.Photo == nil || len(m.Photo) == 0 {
		return nil
	}
	photo := m.Photo[len(m.Photo)-1]
	return File{ID: photo.FileID, UniqueID: photo.FileUniqueID}
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
