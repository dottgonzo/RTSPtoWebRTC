package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/deepch/vdk/codec/h264parser"
	"github.com/deepch/vdk/format/rtsp"
)

func dateToStringSequenceID(datetime time.Time) string {
	// now.Year()+now.Month()+now.Day()+now.Hour()+now.Minute()+now.Second()
	var stringGDateTime = fmt.Sprintf("%02d", datetime.Year()) +
		fmt.Sprintf("%02d", datetime.Month()) +
		fmt.Sprintf("%02d", datetime.Day()) +
		fmt.Sprintf("%02d", datetime.Hour()) +
		fmt.Sprintf("%02d", datetime.Minute()) +
		fmt.Sprintf("%02d", datetime.Second())
	return stringGDateTime
}

func getNewStreamFileName(name string) string {
	var baseName = "stream_" + name
	var sequenceTime = dateToStringSequenceID(time.Now())

	return baseName + "_" + sequenceTime + ".flv"
}
func getNewStreamFilePath(name string) string {
	streamRecordDirEnv, streamRecordDirErr := os.LookupEnv("STREAMDIR")
	if !streamRecordDirErr || streamRecordDirEnv == "" {
		println("using default port config :80")

		streamRecordDirEnv = "/streams"
	}

	return streamRecordDirEnv + "/" + getNewStreamFileName(name)
}
func serveStreams() {

	for k, v := range Config.Streams {
		go func(name, url string) {
			for {

				f, err := os.OpenFile(getNewStreamFilePath(name), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					panic(err)
				}

				defer f.Close()
				log.Println(name, "connect", url)
				rtsp.DebugRtsp = true
				session, err := rtsp.Dial(url)
				if err != nil {
					log.Println(name, err)
					time.Sleep(5 * time.Second)
					continue
				}
				session.RtpKeepAliveTimeout = 10 * time.Second
				if err != nil {
					log.Println(name, err)
					time.Sleep(5 * time.Second)
					continue
				}
				codec, err := session.Streams()
				if err != nil {
					log.Println(name, err)
					time.Sleep(5 * time.Second)
					continue
				}

				Config.coAd(name, codec)
				for {
					// if(time.Now(){
					// 	f, err := os.OpenFile(getNewStreamFilePath(name), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

					// }
					pkt, err := session.ReadPacket()
					if err != nil {
						log.Println(name, err)
						break
					}
					// pkt.Time = time.Duration(codec[0]) * time.Second / time.Duration(stream.timeScale())

					var chunk []byte

					if pkt.IsKeyFrame {

						if pkt.Time.Seconds() > 0 && int(pkt.Time.Seconds())%1800 == 0 {
							f, err = os.OpenFile(getNewStreamFilePath(name), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
							if err != nil {
								panic(err)
							}
							defer f.Close()

						}
						sps := codec[0].(h264parser.CodecData).SPS()
						pps := codec[0].(h264parser.CodecData).PPS()
						chunk = append([]byte{0, 0, 0, 1}, bytes.Join([][]byte{sps, pps, pkt.Data[4:]}, []byte{0, 0, 0, 1})...)
					} else {
						chunk = append([]byte{0, 0, 0, 1}, pkt.Data[4:]...)
					}

					f.Write(chunk)
					// pkt.Time = time.Duration(pkt) * time.Second / time.Duration(stream.timeScale())

					// log.Println(name, "test", sps)
					Config.cast(name, pkt)
				}
				err = session.Close()
				if err != nil {
					log.Println("session Close error", err)
				}
				log.Println(name, "reconnect wait 5s")
				f.Close()
				time.Sleep(5 * time.Second)
			}
		}(k, v.URL)
	}
	log.Println("all streamings started")

	for {
		log.Println("check streaming")

		time.Sleep(5 * time.Second)

	}

}
