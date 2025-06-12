package ihttp

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"
)

func Md5(content string) (md string) {
	h := md5.New()
	h.Write([]byte(content))
	md = fmt.Sprintf("%x", h.Sum(nil))
	return
}

func Test_PublishDrama(t *testing.T) {
	// exporter, _ := tracer.TraceExporterWithStdout()
	// shutdown, _ := tracer.InitProvider(exporter, 0)
	// defer shutdown(context.Background())

	dramaIds := []int64{
		85446598,
	}

	headers := map[string]string{
		"access-token": "a25fdb1d-9f5e-461b-b6bc-d114ca46a9ab",
	}

	for _, dId := range dramaIds {
		params := map[string]int64{
			"id": dId,
		}
		sendData, _ := json.Marshal(params)
		fmt.Println("to publish drama: ", string(sendData))
		resp, _ := PostJson(context.Background(), "https://pre-api.welltop.tech/api/v3/ops/content/series/publish", bytes.NewBuffer(sendData), 3*time.Second, headers)
		defer resp.Body.Close()
		body, er := io.ReadAll(resp.Body)
		fmt.Println("publish response ", string(body), er)
	}
}

type Callback struct {
	PaidAt                  string `json:"paid_at"`
	PastPaymentAttemptCount int32  `json:"past_payment_attempt_count"`
	Status                  string `json:"status"`
	SubscriptionId          string `json:"subscription_id"`
	TotalAmount             int32  `json:"total_amount"`
	Currency                string `json:"currency"`
	Id                      string `json:"id"`
	UpdatedAt               string `json:"updated_at"`
	CustomerId              string `json:"customer_id"`
	CreatedAt               string `json:"created_at"`
}

func Test_ParseAirWallexCallback(t *testing.T) {
	data := `{"id":"inv_sgpdwfhrth86yb8awl9","past_payment_attempt_count":0,"subscription_id":"sub_hkpd6pcfth3h5sxe4ih","total_amount":0,"updated_at":"2025-06-11T09:10:25+0000","created_at":"2025-06-11T09:10:25+0000","currency":"USD","paid_at":"2025-06-11T09:10:26+0000","status":"PAID","customer_id":"cus_hkpd5t2nwh3h5sloica"}
{"currency":"USD","id":"inv_sgpd8wc9bh821tdzm9x","status":"PAID","subscription_id":"sub_sgpd8wc9bh821tdzli3","customer_id":"cus_hkpd5slpwh7ggqfyzrv","created_at":"2025-06-11T10:05:12+0000","paid_at":"2025-06-11T10:05:12+0000","past_payment_attempt_count":0,"total_amount":0,"updated_at":"2025-06-11T10:05:12+0000"}
{"updated_at":"2025-06-11T10:03:36+0000","customer_id":"cus_hkpdwdmmfh7x2hx6uvy","created_at":"2025-06-11T10:03:36+0000","past_payment_attempt_count":0,"status":"PAID","total_amount":0,"currency":"USD","id":"inv_sgpdnjpsph84se4j2fc","paid_at":"2025-06-11T10:03:36+0000","subscription_id":"sub_sgpdvxs67h7x2jtzupa"}
{"created_at":"2025-06-11T07:41:35+0000","currency":"USD","id":"inv_sgpdnjpsph86vv2diun","paid_at":"2025-06-11T07:41:36+0000","status":"PAID","customer_id":"cus_hkpdpg2hnh3h3cbozbw","past_payment_attempt_count":0,"subscription_id":"sub_hkpdrhpgkh3h3cr48xf","total_amount":0,"updated_at":"2025-06-11T07:41:35+0000"}
{"created_at":"2025-06-11T08:11:05+0000","id":"inv_sgpdwfhrth86lnvvppd","paid_at":"2025-06-11T08:11:05+0000","past_payment_attempt_count":0,"status":"PAID","subscription_id":"sub_hkpd49fnnh56jvv7nla","total_amount":0,"currency":"USD","updated_at":"2025-06-11T08:11:05+0000","customer_id":"cus_hkpdhppsth56jvko7ux"}
{"id":"inv_sgpdmbtlhh86xidd7zm","past_payment_attempt_count":0,"status":"PAID","created_at":"2025-06-11T08:41:24+0000","currency":"EUR","paid_at":"2025-06-11T08:41:24+0000","subscription_id":"sub_sgpdqcw8rh6wmgn06xp","total_amount":0,"updated_at":"2025-06-11T08:41:24+0000","customer_id":"cus_hkpdl78z6h6wm55et12"}
{"subscription_id":"sub_sgpd5bd5qh7vv0ksr2l","updated_at":"2025-06-11T09:08:05+0000","currency":"USD","id":"inv_sgpdcwgpjh83kuv9phe","past_payment_attempt_count":0,"status":"PAID","total_amount":0,"customer_id":"cus_hkpd7xwwsh7vv08te5x","created_at":"2025-06-11T09:08:05+0000","paid_at":"2025-06-11T09:08:05+0000"}
{"created_at":"2025-06-11T14:39:26+0000","currency":"USD","past_payment_attempt_count":0,"subscription_id":"sub_hkpd57np8h273tq67j1","total_amount":0,"updated_at":"2025-06-11T14:39:26+0000","id":"inv_sgpdnjpsph877drdsul","paid_at":"2025-06-11T14:39:27+0000","status":"PAID","customer_id":"cus_hkpdlzdzch273rizfp8"}
{"customer_id":"cus_hkpdd2rwch875u68dsl","created_at":"2025-06-11T15:01:31+0000","id":"inv_sgpdwfhrth875vctl95","paid_at":"2025-06-11T15:01:31+0000","past_payment_attempt_count":0,"total_amount":0,"currency":"USD","status":"PAID","subscription_id":"sub_sgpdwfhrth875vctl93","updated_at":"2025-06-11T15:01:31+0000"}
{"id":"inv_sgpdnjpsph878k458ov","past_payment_attempt_count":0,"status":"PAID","updated_at":"2025-06-11T15:22:17+0000","customer_id":"cus_hkpdc2d52h3wustveap","created_at":"2025-06-11T15:22:17+0000","currency":"USD","paid_at":"2025-06-11T15:22:17+0000","subscription_id":"sub_hkpdbhmrvh3wvqdk9af","total_amount":0}
{"status":"PAID","subscription_id":"sub_hkpdrxnlrh649nqaqzt","total_amount":0,"updated_at":"2025-06-11T15:26:55+0000","customer_id":"cus_hkpd77vnth649narisq","created_at":"2025-06-11T15:26:55+0000","currency":"USD","id":"inv_sgpd8wc9bh7ziuivx0z","paid_at":"2025-06-11T15:26:55+0000","past_payment_attempt_count":0}
{"currency":"USD","paid_at":"2025-06-11T15:55:53+0000","past_payment_attempt_count":0,"subscription_id":"sub_sgpdpdjsph71d7amrby","updated_at":"2025-06-11T15:55:53+0000","created_at":"2025-06-11T15:55:53+0000","id":"inv_sgpd8wc9bh7zjn5i425","status":"PAID","total_amount":0,"customer_id":"cus_hkpdl78z6h71cudiv8t"}
{"customer_id":"cus_hkpd7r95kh7k5qaq6ar","created_at":"2025-06-11T17:00:22+0000","paid_at":"2025-06-11T17:00:22+0000","past_payment_attempt_count":0,"status":"PAID","subscription_id":"sub_sgpdr8p8bh7k5qr5pcr","total_amount":0,"currency":"USD","id":"inv_sgpdnjpsph87b9mfyff","updated_at":"2025-06-11T17:00:22+0000"}
{"customer_id":"cus_hkpd5fgh4h7ylr61vbn","created_at":"2025-06-11T19:29:14+0000","currency":"USD","id":"inv_sgpdmbtlhh86boesq1q","updated_at":"2025-06-11T19:29:14+0000","paid_at":"2025-06-11T19:29:15+0000","past_payment_attempt_count":0,"status":"PAID","subscription_id":"sub_sgpds5pfnh7ylu4aebk","total_amount":0}
{"created_at":"2025-06-11T19:36:57+0000","currency":"EUR","paid_at":"2025-06-11T19:36:57+0000","past_payment_attempt_count":0,"status":"PAID","customer_id":"cus_hkpdv7hqch6l1ffq4hk","id":"inv_sgpdcwgpjh7zpqkba5f","subscription_id":"sub_sgpdqcw8rh71japtmky","total_amount":0,"updated_at":"2025-06-11T19:36:57+0000"}
{"id":"inv_sgpdnjpsph87fttb5j8","paid_at":"2025-06-11T19:45:53+0000","past_payment_attempt_count":0,"status":"PAID","subscription_id":"sub_sgpd8wc9bh7zpziufqn","total_amount":0,"customer_id":"cus_hkpdxg6q8h6dbonwmd9","currency":"USD","updated_at":"2025-06-11T19:45:53+0000","created_at":"2025-06-11T19:45:53+0000"}
{"updated_at":"2025-06-11T21:33:49+0000","created_at":"2025-06-11T21:33:49+0000","currency":"USD","id":"inv_sgpdnjpsph87iswd5la","past_payment_attempt_count":0,"customer_id":"cus_hkpdfwwtph7zswums3v","paid_at":"2025-06-11T21:33:49+0000","status":"PAID","subscription_id":"sub_sgpdcwgpjh7zsylgon0","total_amount":0}
{"created_at":"2025-06-11T21:54:23+0000","paid_at":"2025-06-11T21:54:23+0000","past_payment_attempt_count":0,"status":"PAID","subscription_id":"sub_sgpd5bd5qh7t94rh98k","updated_at":"2025-06-11T21:54:23+0000","customer_id":"cus_hkpdwdmmfh7t90nl04b","currency":"USD","id":"inv_sgpd5bd5qh7t94rh98m","total_amount":0}
{"past_payment_attempt_count":0,"status":"PAID","updated_at":"2025-06-11T22:00:21+0000","created_at":"2025-06-11T22:00:21+0000","currency":"USD","id":"inv_sgpdwfhrth86fud2fi2","paid_at":"2025-06-11T22:00:21+0000","subscription_id":"sub_sgpd6hbpnh73umwyrod","total_amount":0,"customer_id":"cus_hkpdhscjnh73ulhcs6d"}
{"currency":"USD","id":"inv_sgpdnjpsph84yzii85q","status":"PAID","subscription_id":"sub_hkpdz54cvh69psq6nad","total_amount":0,"updated_at":"2025-06-11T23:16:16+0000","customer_id":"cus_hkpdw22b5h69ps2tf73","paid_at":"2025-06-11T23:16:16+0000","past_payment_attempt_count":0,"created_at":"2025-06-11T23:16:16+0000"}
{"updated_at":"2025-06-12T01:29:40+0000","created_at":"2025-06-12T01:29:40+0000","currency":"USD","id":"inv_sgpdnjpsph87pb0h8kv","past_payment_attempt_count":0,"customer_id":"cus_hkpdk2s7bh7zzc3736p","paid_at":"2025-06-12T01:29:40+0000","status":"PAID","subscription_id":"sub_sgpd8wc9bh7zzgq0mmg","total_amount":0}
{"created_at":"2025-06-12T02:28:31+0000","id":"inv_sgpdmbtlhh87mufqc0y","paid_at":"2025-06-12T02:28:31+0000","subscription_id":"sub_hkpdqqmvbh62gfk7xh1","currency":"USD","past_payment_attempt_count":0,"status":"PAID","total_amount":0,"updated_at":"2025-06-12T02:28:31+0000","customer_id":"cus_hkpdxnk9ph62gf6xfpe"}
{"currency":"USD","id":"inv_sgpdmbtlhh87rz4u5lz","past_payment_attempt_count":0,"status":"PAID","subscription_id":"sub_hkpdnmj5vh2utxyvezr","created_at":"2025-06-12T03:06:33+0000","paid_at":"2025-06-12T03:06:33+0000","total_amount":0,"updated_at":"2025-06-12T03:06:33+0000","customer_id":"cus_hkpdq8nsbh2utwvx8ml"}
{"paid_at":"2025-06-12T03:23:44+0000","past_payment_attempt_count":0,"status":"PAID","subscription_id":"sub_hkpdk5fvbh58ucxnsd4","total_amount":0,"currency":"USD","id":"inv_sgpd95lwlh829zkugc9","updated_at":"2025-06-12T03:23:44+0000","customer_id":"cus_hkpdpthl2h58u3if21r","created_at":"2025-06-12T03:23:44+0000"}`
	dataSli := strings.Split(data, "\n")
	subIds := make([]string, 0, 50)
	for _, v := range dataSli {
		fmt.Println("v==", v)
		var item Callback
		err := json.Unmarshal([]byte(v), &item)
		if err != nil {
			fmt.Println("json decode err", err)
			return

		}
		subIds = append(subIds, item.SubscriptionId)
	}
	fmt.Println("subIds ", subIds)
	subIdStr := "'" + strings.Join(subIds, "','") + "'"
	fmt.Println("subIdStr==>", subIdStr)
}
