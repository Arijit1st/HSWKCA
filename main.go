package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

type VotesData struct {
	StartTime      time.Time     `json:"startTime"`
	Till           time.Time     `json:"till"`
	TimElapsed     time.Duration `json:"timeElapsed"`
	TotalVotesSent int           `json:"totalVotesSent"`
}

func main() {
	votesData := VotesData{}
	getVotesDataFromJSON, errReadingFile := os.ReadFile("./votesData.json")
	if errReadingFile != nil {
		log.Println(errReadingFile)
		return
	}
	errUnmarshalling := json.Unmarshal(getVotesDataFromJSON, &votesData)
	if errUnmarshalling != nil {
		log.Println(errUnmarshalling)
		return
	}

	log.Printf("%v votes sent successfully starting from %v till %v (%v elapsed)\n\n", votesData.TotalVotesSent, votesData.StartTime.Format(time.DateTime), time.Now().Format(time.DateTime), time.Since(votesData.StartTime))

	for range 1000 {
		go func() {
			for {
				reqData := map[string]any{
					"user_id":     uuid.NewString(),
					"region":      "global",
					"environment": "production",
					"votes": []map[string]string{
						{
							"question_id": "c473c6fb-972d-4a6b-838e-c11a005726b9",
							"option_id":   "5c68618e-40ab-4b30-a348-33dbd9141148",
						},
					},
				}

				reqDataJSON, errMarshalling := json.Marshal(reqData)
				if errMarshalling != nil {
					log.Println(errMarshalling)
					continue
				}

				res, errRes := http.NewRequest("POST", "https://prod.kca.engage.mik.studio/api/vote", bytes.NewReader(reqDataJSON))
				if errRes != nil {
					log.Println(errRes)
					continue
				}
				res.Header.Set("Content-Type", "application/json")
				res.Header.Set("X-Forwarded-For", fmt.Sprintf("%v.%v.%v.%v", rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(256)))

				resSend, errSend := http.DefaultClient.Do(res)
				if errSend != nil {
					log.Println(errSend)
					continue
				}

				if resSend.StatusCode == 200 {
					votesData.TotalVotesSent++
					votesData.Till = time.Now()
					votesData.TimElapsed = time.Since(votesData.StartTime)
					marshallData, errMarshallingFileData := json.Marshal(votesData)
					if errMarshallingFileData != nil {
						log.Println(errMarshallingFileData)
						return
					}
					errWritingFile := os.WriteFile("./votesData.json", marshallData, 0o666)
					if errWritingFile != nil {
						log.Println(errWritingFile)
						return
					}
				}

				rawResData, errReadingRawResData := io.ReadAll(resSend.Body)

				if errReadingRawResData != nil {
					log.Println(errReadingRawResData)
					return
				}

				fmt.Println(string(rawResData))

				resData := make([]map[string]any, 0)
				errUnmarshallingResData := json.Unmarshal(rawResData, &resData)
				if errUnmarshallingResData != nil {
					log.Println(errUnmarshallingResData)
					return
				}

				resDataFormatted, errFormatting := json.MarshalIndent(resData, "", "    ")
				if errFormatting != nil {
					log.Println(errFormatting)
					return
				}
				log.Printf("%v\n\n", string(resDataFormatted))

				log.Printf("%v votes sent successfully starting from %v till %v (%v elapsed)\n\n", votesData.TotalVotesSent, votesData.StartTime.Format(time.DateTime), time.Now().Format(time.DateTime), time.Since(votesData.StartTime))

				errClosingBody := resSend.Body.Close()
				if errClosingBody != nil {
					log.Println(errClosingBody)
				}
			}
		}()
	}
	select {}
}
