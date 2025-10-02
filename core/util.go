package core

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"hash/crc32"
	"net/url"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/star-39/moe-sticker-bot/pkg/msbimport"
	tele "gopkg.in/telebot.v3"
	"mvdan.cc/xurls/v2"
)

var regexAlphanum = regexp.MustCompile(`[a-zA-Z0-9_]+`)

// var httpClient = &http.Client{
// 	Timeout: 5 * time.Second,
// }

func checkTitle(t string) bool {
	if len(t) > 128 || len(t) < 1 {
		return false
	} else {
		return true
	}
}

func checkID(s string) bool {
	maxL := 64 - len(botName)
	if len(s) < 1 || len(s) > maxL {
		return false
	}
	if _, err := strconv.Atoi(s[:1]); err == nil {
		return false
	}
	if strings.Contains(s, "__") {
		return false
	}
	if strings.Contains(s, " ") {
		return false
	}
	//Telegram does not allow sticker name having the word "telegram"
	if strings.Contains(s, "telegram") {
		return false
	}

	return true
}

func secHex(n int) string {
	bytes := make([]byte, n)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// func secNum(n int) string {
// 	numbers := ""
// 	for i := 0; i < n; i++ {
// 		randInt, _ := rand.Int(rand.Reader, big.NewInt(10))
// 		numbers += randInt.String()
// 	}
// 	return numbers
// }

func findLink(s string) string {
	rx := xurls.Strict()
	return rx.FindString(s)
}

func findLinkWithType(s string) (string, string) {
	rx := xurls.Strict()
	link := rx.FindString(s)
	if link == "" {
		return "", ""
	}

	u, _ := url.Parse(link)
	host := u.Host

	if host == "t.me" {
		host = LINK_TG
	} else if strings.HasSuffix(host, "line.me") {
		host = LINK_IMPORT
	} else if strings.HasSuffix(host, "kakao.com") {
		host = LINK_IMPORT
	}

	log.Debugf("link found within findLinkWithType: link=%s, host=%s", link, host)
	return link, host
}

func findEmojis(s string) string {
	out, err := exec.Command("msb_emoji.py", "string", s).Output()
	if err != nil {
		return ""
	}
	return string(out)
}

func findEmojiList(s string) []string {
	out, err := exec.Command("msb_emoji.py", "json", s).Output()
	if err != nil {
		return []string{}
	}
	list := []string{}
	json.Unmarshal(out, &list)
	return list
}

func stripEmoji(s string) string {
	out, err := exec.Command("msb_emoji.py", "text", s).Output()
	if err != nil {
		return ""
	}
	return string(out)
}

func sanitizeCallback(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		log.Debug("Sanitizing callback data...")
		c.Callback().Data = regexAlphanum.FindString(c.Callback().Data)

		log.Debugln("now:", c.Callback().Data)
		return next(c)
	}
}
func autoRespond(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		if c.Callback() != nil {
			defer c.Respond()
		}
		return next(c)
	}
}

func escapeTagMark(s string) string {
	s = strings.ReplaceAll(s, "<", "＜")
	s = strings.ReplaceAll(s, ">", "＞")
	return s
}

func getSIDFromMessage(m *tele.Message) string {
	if m.Sticker != nil {
		return m.Sticker.SetName
	}

	link := findLink(m.Text)
	return path.Base(link)
}

func retrieveSSDetails(c tele.Context, id string, sd *StickerData) error {
	ss, err := c.Bot().StickerSet(id)
	if err != nil {
		return err
	}
	sd.stickerSet = ss
	sd.title = ss.Title
	sd.id = ss.Name
	sd.cAmount = len(ss.Stickers)
	sd.stickerSetType = ss.Type
	if ss.Type == tele.StickerCustomEmoji {
		sd.isCustomEmoji = true
	}
	return nil
}

func GetUd(uidS string) (*UserData, error) {
	uid, err := strconv.ParseInt(uidS, 10, 64)
	if err != nil {
		return nil, err
	}
	ud, ok := users.data[uid]
	if ok {
		return ud, nil
	} else {
		return nil, errors.New("no such user in state")
	}
}

func sliceMove[T any](oldIndex int, newIndex int, slice []T) []T {
	orig := slice
	element := slice[oldIndex]

	if oldIndex > newIndex {
		if len(slice)-1 == oldIndex {
			slice = slice[0 : len(slice)-1]
		} else {
			slice = append(slice[0:oldIndex], slice[oldIndex+1:]...)
		}
		slice = append(slice[:newIndex], append([]T{element}, slice[newIndex:]...)...)
	} else if oldIndex < newIndex {
		slice = append(slice[0:oldIndex], slice[oldIndex+1:]...)
		if newIndex != len(slice) {
			newIndex = newIndex + 1
		}
		slice = append(slice[:newIndex], append([]T{element}, slice[newIndex:]...)...)
	} else {
		return orig
	}
	return slice
}

func chunkSlice(slice []string, chunkSize int) [][]string {
	var chunks [][]string
	for {
		if len(slice) == 0 {
			break
		}

		if len(slice) < chunkSize {
			chunkSize = len(slice)
		}

		chunks = append(chunks, slice[0:chunkSize])
		slice = slice[chunkSize:]
	}
	return chunks
}

func compCRC32(f1 string, f2 string) bool {
	fb1, err := os.ReadFile(f1)
	if err != nil {
		return false
	}
	fb2, err := os.ReadFile(f2)
	if err != nil {
		return false
	}

	c1 := crc32.ChecksumIEEE(fb1)
	c2 := crc32.ChecksumIEEE(fb2)
	log.Debugf("File:%s, C:%v", f1, c1)
	log.Debugf("File:%s, C:%v", f2, c2)

	if c1 == c2 {
		return true
	} else {
		return false
	}
}

// func hashCRC64(s string) string {
// 	h := crc64.New(crc64.MakeTable(crc64.ISO))
// 	h.Write([]byte(s))
// 	csum := fmt.Sprintf("%x", h.Sum(nil))
// 	return csum
// }

func checkGnerateSIDFromLID(ld *msbimport.LineData) string {
	id := ld.Id
	id = strings.ReplaceAll(id, "-", "_")
	id = strings.ReplaceAll(id, "__", "_")

	s := ld.Store + id + secHex(2) + "_by_" + botName

	if len(s) > 64 {
		log.Debugln("id too long:", len(s))
		extra := len(s) - 64
		id = id[:len(id)-extra]
		s = ld.Store + id + secHex(2) + "_by_" + botName
		s = strings.ReplaceAll(s, "__", "_")
		log.Debugln("Shortend id to:", s)
	}

	return s
}

// // Local bot api returns a absolute path in FilePath.
// // We need to separate "real" api server and local api server.
// // We move the file from api server to target location.
// // Be careful, this does not work when crossing mount points.
// func teleDownload(tf *tele.File, f string) error {
// 	// if msbconf.BotApiAddr != "" {
// 	// 	tf2, err := b.FileByID(tf.FileID)
// 	// 	if err != nil {
// 	// 		return err
// 	// 	}
// 	// 	err = os.Rename(tf2.FilePath, f)
// 	// 	if err != nil {
// 	// 		exec.Command("cp", tf2.FilePath, f).CombinedOutput()
// 	// 	}
// 	// 	return os.Chmod(f, 0644)
// 	// } else {
// 	return b.Download(tf, f)
// 	// }
// }

// To comply with new InputSticker requirement on format,
// guess format based on file extension.
func guessInputStickerFormat(f string) string {
	if strings.HasSuffix(f, ".webm") {
		return "video"
	} else {
		return "static"
	}
}
