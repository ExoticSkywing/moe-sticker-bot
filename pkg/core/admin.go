package core

import (
	"strings"

	tele "gopkg.in/telebot.v3"
)

// import (
// 	"os"
// 	"path/filepath"
// 	"strconv"
// 	"strings"

// 	log "github.com/sirupsen/logrus"
// 	tele "gopkg.in/telebot.v3"
// )

// // DANGER ZONE!
// // DO NOT USE UNLESS YOU SCRUTINIZED THE CODE.

// // This command is to sanitize duplicated sticker in a set, or update its auto_emoji status.
// // You should not use this command unless you were using the python version before.
// // It takes forever to run for HUGE databases.
// func cmdSanitize(c tele.Context) error {
// 	if ADMIN_UID != c.Sender().ID {
// 		return c.Send("Admin only command. /start")
// 	}

// 	msgText := c.Message().Text
// 	args := strings.Split(msgText, " ")
// 	if len(args) <= 1 {
// 		return c.Send("Missing subcommand! invalid / dup / all / ae")
// 	}
// 	startIndex, _ := strconv.Atoi(args[2])
// 	switch args[1] {
// 	case "invalid":
// 		sanitizeInvalidSSinDB(startIndex, c)
// 	case "ae":
// 		sanitizeAE(startIndex, c)
// 	default:
// 		sanitizeDatabase(startIndex, c)
// 	}
// 	return nil
// }

// func sanitizeAE(startIndex int, c tele.Context) error {
// 	c.Send("Started.")
// 	ls := queryLineS("QUERY_ALL")
// 	for i, l := range ls {
// 		if i < startIndex {
// 			continue
// 		}
// 		log.Infof("Checking:%s", l.Tg_id)
// 		ss, err := c.Bot().StickerSet(l.Tg_id)
// 		if err != nil {
// 			if strings.Contains(err.Error(), "is invalid") {
// 				log.Infof("SS:%s is invalid. purging it from db...", l.Tg_id)
// 				go c.Send("purging invalid: https://t.me/addstickers/" + l.Tg_id)
// 				deleteLineS(l.Tg_id)
// 				deleteUserS(l.Tg_id)
// 			} else {
// 				c.Send("Unknow error? " + err.Error())
// 				log.Errorln(err)
// 			}
// 			continue
// 		}
// 		for si := range ss.Stickers {
// 			if si > 0 {
// 				if ss.Stickers[si].Emoji != ss.Stickers[si-1].Emoji {
// 					log.Warnln("Setting auto emoji to FALSE for ", l.Tg_id)
// 					updateLineSAE(false, l.Tg_id)
// 				}
// 			}
// 		}
// 	}
// 	c.Send("Sanitize AE done!")
// 	return nil
// }

// func sanitizeInvalidSSinDB(startIndex int, c tele.Context) error {
// 	msg, _ := c.Bot().Send(c.Recipient(), "0")
// 	ls := queryLineS("QUERY_ALL")
// 	log.Infoln(ls)
// 	for i, l := range ls {
// 		if i < startIndex {
// 			continue
// 		}
// 		log.Infof("Checking:%s", l.Tg_id)
// 		_, err := c.Bot().StickerSet(l.Tg_id)
// 		if err != nil {
// 			if strings.Contains(err.Error(), "is invalid") {
// 				log.Warnf("SS:%s is invalid. purging it from db...", l.Tg_id)
// 				go c.Send("purging: https://t.me/addstickers/" + l.Tg_id)
// 				deleteLineS(l.Tg_id)
// 				deleteUserS(l.Tg_id)
// 			} else {
// 				go c.Send("Unknow error? " + err.Error())
// 				log.Errorln(err)
// 			}
// 		}
// 		go c.Bot().Edit(msg, "line sanitize invalid: "+strconv.Itoa(i))
// 	}
// 	us := queryUserS(-1)
// 	log.Infoln(us)
// 	for i, u := range us {
// 		log.Infof("Checking:%s", u.tg_id)
// 		_, err := c.Bot().StickerSet(u.tg_id)
// 		if err != nil {
// 			if strings.Contains(err.Error(), "is invalid") {
// 				log.Warnf("SS:%s is invalid. purging it from db...", u.tg_id)
// 				go c.Send("purging: https://t.me/addstickers/" + u.tg_id)
// 				deleteUserS(u.tg_id)
// 			} else {
// 				go c.Send("Unknow error? " + err.Error())
// 				log.Errorln(err)
// 			}
// 		}
// 		go c.Bot().Edit(msg, "user S sanitize invalid: "+strconv.Itoa(i))
// 	}
// 	c.Send("Sanitize invalid done!")
// 	return nil
// }

// func sanitizeDatabase(startIndex int, c tele.Context) error {
// 	msg, _ := c.Bot().Send(c.Recipient(), "0")
// 	ls := queryLineS("QUERY_ALL")
// 	log.Infoln(ls)
// 	for i, l := range ls {
// 		if i < startIndex {
// 			continue
// 		}
// 		log.Debugf("Scanning:%s", l.Tg_id)
// 		ss, err := c.Bot().StickerSet(l.Tg_id)
// 		if err != nil {
// 			if strings.Contains(err.Error(), "is invalid") {
// 				log.Infof("SS:%s is invalid. purging it from db...", l.Tg_id)
// 				go c.Send("purging invalid: https://t.me/addstickers/" + l.Tg_id)
// 				deleteLineS(l.Tg_id)
// 				deleteUserS(l.Tg_id)
// 			} else {
// 				c.Send("Unknow error? " + err.Error())
// 				log.Errorln(err)
// 			}
// 			continue
// 		}
// 		workdir := filepath.Join(dataDir, secHex(8))
// 		os.MkdirAll(workdir, 0755)
// 		for si, s := range ss.Stickers {
// 			if si > 0 {
// 				if ss.Stickers[si].Emoji != ss.Stickers[si-1].Emoji {
// 					log.Warnln("Setting auto emoji to FALSE for ", l.Tg_id)
// 					updateLineSAE(false, l.Tg_id)
// 				}
// 			}

// 			fp := filepath.Join(workdir, strconv.Itoa(si-1)+".webp")
// 			f := filepath.Join(workdir, strconv.Itoa(si)+".webp")
// 			c.Bot().Download(&s.File, f)

// 			if compCRC32(f, fp) {
// 				c.Bot().DeleteSticker(s.FileID)
// 				log.Warnf("Deleted on animated dup s!")
// 				c.Send("Deleted dup S from: https://t.me/addstickers/" + s.SetName + "  indexis: " + strconv.Itoa(si))
// 			}

// 		}
// 		os.RemoveAll(workdir)

// 		go c.Bot().Edit(msg, "line s sanitize all: "+strconv.Itoa(i))
// 	}
// 	c.Send("ALL SANITIZED!")
// 	return nil
// }

func cmdStatRep(c tele.Context) error {
	// Report status.
	stat := []string{}
	py_emoji_ok, _ := httpGet("http://127.0.0.1:5000/status")
	stat = append(stat, "py_emoji_ok? :"+py_emoji_ok)
	return c.Send(strings.Join(stat, "\n"))
}
