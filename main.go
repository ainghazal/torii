package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

var (
	authorName = "Ain Ghazal <ain@openobservatory.org>"
)

type config struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Author      string    `json:"author"`
	NetTests    []netTest `json:"nettests"`
}

type netTest struct {
	TestName string   `json:"test_name"`
	Inputs   []string `json:"inputs"`
	Options  options  `json:"options"`
}

type options struct {
	Cipher   string
	Auth     string
	SafeCa   string
	SafeCert string
	SafeKey  string
}

var (
	ca   = `base64:LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJZakNDQVFpZ0F3SUJBZ0lCQVRBS0JnZ3Foa2pPUFFRREFqQVhNUlV3RXdZRFZRUURFd3hNUlVGUUlGSnYKYjNRZ1EwRXdIaGNOTWpFeE1UQXlNVGt3TlRNM1doY05Nall4TVRBeU1Ua3hNRE0zV2pBWE1SVXdFd1lEVlFRRApFd3hNUlVGUUlGSnZiM1FnUTBFd1dUQVRCZ2NxaGtqT1BRSUJCZ2dxaGtqT1BRTUJCd05DQUFReE9YQkd1K2dmCnBqSHpWdGVHVFdMNlhuRnh0RW5LTUZwS2FKa0EvVk9IbUVTem9Mc1pSUXh0ODhHc3N4YXFDMDFKMTdpZFFpcXYKemdOcGVkbXR2RnR5bzBVd1F6QU9CZ05WSFE4QkFmOEVCQU1DQXFRd0VnWURWUjBUQVFIL0JBZ3dCZ0VCL3dJQgpBVEFkQmdOVkhRNEVGZ1FVWmRvVWxKckNJVU5GcnBmZkFxK0xRam53RXo0d0NnWUlLb1pJemowRUF3SURTQUF3ClJRSWdmcjN3NHRuUkcrTmRJM0xzR1Bsc1JrdEdLMjB4SFR6c0Izb3JCMHlDNmNJQ0lRQ0IrLzl5OG5tU1N0Zk4KVlVNVXlrMmhOZDcva0M4bkwyMjJUVEQ3VlpVdHNnPT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=`
	key  = `base64:LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFb3dJQkFBS0NBUUVBcTBkVlVnMWRLaWFBMDRmcEtycEYzYlZCTEpsbTdoME4rc1ZIZHA2MVZPZG5iQlBPCmtiMHU5TkJqTFZZaFFzbm1HVWp0TzNHeEdGSTltUEdjN3VHN0wwMkk2aWtIKzBhMWpDNGwyWHVGdEFRWm1xNSsKUWxrTTh5M3FUSzB3NlRPZG9IZUh0TitPaEtyemN2eFJuWXR1ZURqekVOaFl5MkE2UHBTM2w2WWg1a1dxMmZabApxcy9UODhRa2dSSWk5YmxTa2s3RGtKcHFDSXRTVE9zS2RraXJVYTU1d1E2N1dUYzB1ZWQ4YVRONk1mWVJOYWZ4ClRld3Zid3RFMWpEN0Y2SzRQRks3eGVMS2NSMVFlRzA4N0JkVS95N2dIR3MxWnArUGdBaHJIREdqOXZlQkRWUnAKM0dLUjhnNitXMWMzMVh4UWdIMEwwRzVmNjRyZVVGb1pZQkZCSXdJREFRQUJBb0lCQUNWRjhzVldieTNiRHpINQpZNzZPcHVHbXJqWThjKy9obHNjNTQyRm5ER01ic0tBT2QyZXoyZUlnNzFSUWFCQ1d5Mkk1UXBjckdMVUlRS3RsCitSYnJQTWNBZ29raXdML29GVjRhTk5adFVSMXB2d0N3ZEgyUHo0ZWtPRmJUWWM5K0VoRjNzYXFrOCtqZkl2ZWsKL1VYaHIvcXR1Z2V5YlRCbEVvZkg2V1F4SFROMUhwOUNEc2Y0SWRyMmkrSk53TFJjVzhVZXMvNHZ3WWhsRGsrbwpnNDV4NWtNWmtscUlHVUxHbUhSRjVUY2N2RklWYnZ3UmdIdXZuNlNqNnhDdFJvbi9SZG5LSnJ4R25NU0VGbnJYCjdxOTFRSUFmaldIYm10Tk9TYVl1NW9FejdnbTE1R2JmdmlEVXRUNjR3djJJREF1dUY4S09hOGowUWNvV2NMM28KaHpMenpORUNnWUVBelZsRTgxR3o4NVVYa2VtSEdwaEhIdlc2ZG93dE1xUlVTb3FUMkgwSCthK0RlaExTZEhGUgozUWNKbTBPeGpQMU1MRkZiM016cmhTT1pZTHVsTHFWeDF1eFJwYWo1ak4zRVFHNHBxWmI0cnhCcCs4RHM2RCsxCk5TM1VKb1FOamQxZEplemEzbTVuRkdULzJuMWdCRmI1MmRmNmxibSs2TGtJNFA0R1F4UDR5TmNDZ1lFQTFZYTIKU3R1b3lSWlVSOE1QVnltZnVFNVBTQ2VtRE90bStzQXIvSGRPVXJKdTkyditHa1JYQTdiR0EvTjI0VE5XKzViYQpZRTF5Nk9ZYnM0NklGRUlmOERFWU8zRGJKaTlIeHU2S3hkcXExanFXamZVcGJDY2xrZXZQQk15Q3V6VnRLQWNwCnVNdmhkanpidlFTMzlDTGtNVGQray9maC83ZWVqUHhLRFhMSEJKVUNnWUFNd2tjdWR4MGZQVnhCaktrQVZnWFYKUHA5ZlRrWmdweVUxbkhhak5PR1IrZjNKVC9JVG1oYmtETlBqK2NqR1lkYWh5a3hTNDhpZWRSL0tpdDR3ajhjSworNVAzSHhDaVdBVWhtN2FxK3Q1b3dqUlRtQ0VnTFJVdFFMTzEwTzZtcWVKbndOZTRpbE9OU05rODBoMXRKNXBPCmxzVFRHTDlyNWxOTzUzbXNJVW1MOFFLQmdBeDRBRjhnc3B1RGZVcHZmbzdWZEdrNzBXOWlPVlVaemZxb2pDa0QKQW9UYnZKVWdMa2QwWkN4b1dPblVKc1lCekh1R2xKdjVDZFBGMUNwSkVYTTFaVTRPWDk3Z3VUdGltV3RwZEpzWApLTkMzdlNEdkJ3czB3Z0hpWmtWZWQrZmN0OUlWa1A4a2tMYnAyTjhSem5nb0xYRWVUM3J1aDdqNkRQMG9vbDVrCnJIQjlBb0dCQUtxUERHVU9leVNGeTdiNjQyMnBvTkdGdDV4cHBGME1NQ1lJc3MvdzEzRW81dGxMb0U4OEpvbWwKVHlyMW5GVzZ0dkUvd1crYmZ0RU8rWCt4aDdibzNDNHNIeWpUTXB1WnNob2FXOHRrOFl3ZFI5eTNUV1Vxd3BwaQp2b2U0M0E1LzBuNzE5VFVFZ00vOUtoQVB4RzdQYlZDVDh1azF6a1hmNHZXOUpHUTNaaElDCi0tLS0tRU5EIFJTQSBQUklWQVRFIEtFWS0tLS0tCg==`
	cert = `base64:LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNlRENDQWg2Z0F3SUJBZ0lSQU9TcHNFNU5jdDcwTDNjdVlhcllidkl3Q2dZSUtvWkl6ajBFQXdJd016RXgKTUM4R0ExVUVBd3dvVEVWQlVDQlNiMjkwSUVOQklDaGpiR2xsYm5RZ1kyVnlkR2xtYVdOaGRHVnpJRzl1YkhraApLVEFlRncweU1qQTRNREV4TVRJek1qSmFGdzB5TWpBNU1EVXhNVEl6TWpKYU1CUXhFakFRQmdOVkJBTVRDVlZPClRFbE5TVlJGUkRDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBS3RIVlZJTlhTb20KZ05PSDZTcTZSZDIxUVN5Wlp1NGREZnJGUjNhZXRWVG5aMndUenBHOUx2VFFZeTFXSVVMSjVobEk3VHR4c1JoUwpQWmp4bk83aHV5OU5pT29wQi90R3RZd3VKZGw3aGJRRUdacXVma0paRFBNdDZreXRNT2t6bmFCM2g3VGZqb1NxCjgzTDhVWjJMYm5nNDh4RFlXTXRnT2o2VXQ1ZW1JZVpGcXRuMlphclAwL1BFSklFU0l2VzVVcEpPdzVDYWFnaUwKVWt6ckNuWklxMUd1ZWNFT3UxazNOTG5uZkdremVqSDJFVFduOFUzc0wyOExSTll3K3hlaXVEeFN1OFhpeW5FZApVSGh0UE93WFZQOHU0QnhyTldhZmo0QUlheHd4by9iM2dRMVVhZHhpa2ZJT3ZsdFhOOVY4VUlCOUM5QnVYK3VLCjNsQmFHV0FSUVNNQ0F3RUFBYU5uTUdVd0RnWURWUjBQQVFIL0JBUURBZ2VBTUJNR0ExVWRKUVFNTUFvR0NDc0cKQVFVRkJ3TUNNQjBHQTFVZERnUVdCQlJ4UEdnSkMyVXBNYzNPb1E5aVdaQlh6eVpJSnpBZkJnTlZIU01FR0RBVwpnQlI5U21MWS95dEp4SG0yb3JIY2pqNWpCMXlvL2pBS0JnZ3Foa2pPUFFRREFnTklBREJGQWlFQXpWRE9SVUNKCnVScHlFVEZSc1BHcFZEeWI5Mi9PRGkyNS9KZ25RbmloZCtZQ0lBSnI1S0MvTjVkZ1JpRDNmRGV1Q3dPZXdWY0EKaWQ3U3JxU1gvRjZZc2g3RgotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==`
)

type remote struct {
	IP        string
	Port      string
	Transport string
}

var providerRemotes map[string][]remote

func initProviderRemotes() {
	providerRemotes = make(map[string][]remote)

	// TODO fetch this from the provider API
	riseupRemotes := []remote{
		remote{"204.13.164.252", "1194", "udp"},
		remote{"204.13.164.252", "1194", "tcp"},
	}

	providerRemotes["riseup"] = riseupRemotes
}

func pickRandomRemote(provider string) remote {
	all := providerRemotes[provider]
	pick := rand.Intn(len(all))
	return all[pick]
}

func singleConfig(provider string) *config {

	r := pickRandomRemote(provider)

	test := netTest{
		TestName: "openvpn",
		Inputs: []string{
			fmt.Sprintf(
				"vpn://%s.openvpn/?addr=%s:%s&transport=%s",
				provider,
				r.IP,
				r.Port,
				r.Transport,
			)},
		Options: options{
			Cipher: "AES-256-GCM",
			Auth:   "SHA512",
			// TODO pick auth from in-memory store for each provider
			SafeCa:   ca,
			SafeCert: cert,
			SafeKey:  key,
		},
	}
	return &config{
		Name:        fmt.Sprintf("openvpn-%s", provider),
		Description: fmt.Sprintf("measure vpn connection to random %s gateways", provider),
		Author:      authorName,
		NetTests:    []netTest{test},
	}
}

func riseupDescriptor(w http.ResponseWriter, r *http.Request) {
	cfg := singleConfig("riseup")
	json.NewEncoder(w).Encode(cfg)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	initProviderRemotes()
	router := mux.NewRouter().StrictSlash(false)
	router.HandleFunc("/vpn/riseup.json", riseupDescriptor)
	log.Fatal(http.ListenAndServe(":8080", router))
}
