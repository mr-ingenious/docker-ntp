package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type HttpServerStatus struct {
	IsError      bool
	Text         string
	HttpResponse int
	Err          error
}

type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type TrackingResult struct {
	Info [13]KeyValue `json:"info"`
}

type ServerStatsResult struct {
	Info [17]KeyValue `json:"info"`
}

type SourceStatus struct {
	Name          string `json:"name"`
	Np            string `json:"np"`
	Nr            string `json:"nr"`
	Span          string `json:"span"`
	Frequency     string `json:"frequency"`
	FrequencySkew string `json:"frequency_skew"`
	Offset        string `json:"offset"`
	StdDev        string `json:"std_dev"`
}

type SourcesResult struct {
	Sources []Source `json:"sources"`
}

type Source struct {
	Name       string `json:"name"`
	Ms         string `json:"ms"`
	Stratum    string `json:"stratum"`
	Poll       string `json:"poll"`
	Reach      string `json:"reach"`
	LastRx     string `json:"last_rx"`
	LastSample string `json:"last_sample"`
}

/*
The hostname of the client.
The number of NTP packets received from the client.
The number of NTP packets dropped to limit the response rate.
The average interval between NTP packets.
The average interval between NTP packets after limiting the response rate.
Time since the last NTP packet was received
The number of command packets or NTS-KE connections received/accepted from the client.
The number of command packets or NTS-KE connections dropped to limit the response rate.
The average interval between command packets or NTS-KE connections.
Time since the last command packet or NTS-KE connection was received/accepted.
*/
type Client struct {
	Hostname                       string `json:"hostname"`
	NtpPacketsReceived             string `json:"ntp_received"`
	NtpPacketsDropped              string `json:"ntp_dropped"`
	NtpAverageInterval             string `json:"ntp_interval"`
	NtpAverageIntervalAfterRRLimit string `json:"ntp_interval_rr"`
	NtpTimeSinceLastReceived       string `json:"ntp_last_received"`
	NtsKEReceived                  string `json:"ntske_received"`
	NtsKEDropped                   string `json:"ntske_dropped"`
	NtsKEAverageInterval           string `json:"ntske_interval"`
	NtsKETimeSinceLastReceived     string `json:"ntske_last_received"`
}

type ClientsResult struct {
	Clients []Client `json:"clients"`
}

type SourceStatsResult struct {
	SourceStats []SourceStatus `json:"status"`
}

func chrony_info(option string) []byte {
	fmt.Println("webserver: requested chrony info, param ", option)
	out, err := exec.Command("/usr/bin/chronyc", option).Output()

	if err != nil {
		log.Fatal(err)
	}

	return out
}

func chrony_tracking(w http.ResponseWriter, req *http.Request) {
	out := chrony_info("tracking")

	scanner := bufio.NewScanner(strings.NewReader(string(out)))

	var result [13]KeyValue

	i := 0
	for scanner.Scan() {
		line := scanner.Text()
		kv := strings.Split(line, ": ")
		key := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ToLower(strings.TrimSpace(kv[0])), " ", "_"), ")", ""), "(", "")
		value := strings.TrimSpace(kv[1])

		pair := KeyValue{
			Key:   key,
			Value: value,
		}

		result[i] = pair

		i++

		if i > 12 {
			break
		}
	}

	data := &TrackingResult{
		Info: result,
	}

	// b, err := json.MarshalIndent(data, "", "  ")
	b, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("error occurred: %v\n", err)
	}

	fmt.Fprintf(w, "%s", b)
}

/*
	Example output sourcestats:

Name/IP Address            NP  NR  Span  Frequency  Freq Skew  Offset  Std Dev
==============================================================================
80.153.195.191             19  11  327m     +0.105      0.164  -2921us   895us
130.61.133.198             17  10  275m     +0.219      0.289   -527us  1506us
t2.ipfu.de                 64  32   18h     -0.030      0.088   +199us  3700us
nbg01.muxx.net             16   6  258m     +0.055      0.403  -1518us  1707us
ntp3.rrze.uni-erlangen.de   7   3  103m     +0.271      3.843   +384us  3365us
ntp2.rrze.uni-erlangen.de  44  19   12h     +0.013      0.099  -1216us  2221us
ntp1.rrze.uni-erlangen.de  11   5  171m     +0.119      0.845  -1328us  1922us
ntp0.rrze.uni-erlangen.de  43  17   10h     +0.012      0.108  -1740us  2169us
time.fu-berlin.de          40  19   11h     +0.021      0.122  -2928us  2654us
zeit.fu-berlin.de          27  17  448m     +0.041      0.236  -2282us  2650us
gw-001.oit.one             18  11  233m     -0.040      0.348  -3816us  1425us
ptbtime1.ptb.de            16   7  259m     +0.082      0.318   +627us  1356us
ptbtime2.ptb.de             9   5   69m     +2.476      1.321  +4381us   978us
ptbtime3.ptb.de             7   5  103m     +1.383      3.438  +3986us  2732us
193.134.29.11              42  21   11h     -0.013      0.100  +1197us  2300us
*/
func chrony_sourcestats(w http.ResponseWriter, req *http.Request) {
	out := chrony_info("sourcestats")

	scanner := bufio.NewScanner(strings.NewReader(string(out)))

	var states []SourceStatus

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "IP Address") || strings.Contains(line, "==============================================================================") {
			continue
		}

		kv := strings.Fields(line)

		if len(kv) == 8 {
			item := SourceStatus{
				Name:          kv[0],
				Np:            kv[1],
				Nr:            kv[2],
				Span:          kv[3],
				Frequency:     kv[4],
				FrequencySkew: kv[5],
				Offset:        kv[6],
				StdDev:        kv[7],
			}

			states = append(states, item)
		}
	}

	result := &SourceStatsResult{
		SourceStats: states,
	}

	b, err := json.Marshal(result)
	// b, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("error occurred: %v\n", err)
	}

	fmt.Fprintf(w, "%s", b)
}

/*
MS Name/IP address         Stratum Poll Reach LastRx Last sample
===============================================================================
^- 80.153.195.191                2  10   377   763  -6990us[-8147us] +/-   27ms
^+ 130.61.133.198                2  10   377   215  -5208us[-6371us] +/-   11ms
^+ t2.ipfu.de                    3  10   377   530  +1701us[ +541us] +/-   20ms
^+ nbg01.muxx.net                2  10   377   531  -2097us[-3257us] +/-   13ms
^+ ntp3.rrze.uni-erlangen.de     1  10   377   430  -3659us[-4820us] +/-   14ms
^* ntp2.rrze.uni-erlangen.de     1  10   377    90  -1780us[-2944us] +/-   10ms
^+ ntp1.rrze.uni-erlangen.de     1  10   377   969  -2740us[-3894us] +/-   11ms
^+ ntp0.rrze.uni-erlangen.de     1  10   377   697  -6616us[-7774us] +/-   14ms
^- time.fu-berlin.de             1  10   377    54  -3162us[-3162us] +/-   70ms
^- zeit.fu-berlin.de             1  10   377   300  -6445us[-7607us] +/-   67ms
^- gw-001.oit.one                2  10   377   386  -4793us[-5954us] +/-   46ms
^+ ptbtime1.ptb.de               1  10   377   748  -3677us[-4834us] +/-   16ms
^+ ptbtime2.ptb.de               1   8   377    24   -695us[ -695us] +/-   12ms
^+ ptbtime3.ptb.de               1  10   377  1015   -929us[-2083us] +/-   11ms
^+ 193.134.29.11                 1  10   377   947  +2716us[+1562us] +/-   15ms
*/
func chrony_sources(w http.ResponseWriter, req *http.Request) {
	out := chrony_info("sources")

	scanner := bufio.NewScanner(strings.NewReader(string(out)))

	var sources []Source

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "Stratum") || strings.Contains(line, "===========================================") {
			continue
		}

		kv := strings.Fields(line)

		if len(kv) == 9 {
			item := Source{
				Ms:         kv[0],
				Name:       kv[1],
				Stratum:    kv[2],
				Poll:       kv[3],
				Reach:      kv[4],
				LastRx:     kv[5],
				LastSample: kv[6] + " " + kv[7] + " " + kv[8],
			}

			sources = append(sources, item)
		}
		if len(kv) == 10 {
			item := Source{
				Ms:         kv[0],
				Name:       kv[1],
				Stratum:    kv[2],
				Poll:       kv[3],
				Reach:      kv[4],
				LastRx:     kv[5],
				LastSample: kv[6] + " " + kv[7] + " " + kv[8] + " " + kv[9],
			}

			sources = append(sources, item)
		}
	}

	result := &SourcesResult{
		Sources: sources,
	}

	b, err := json.Marshal(result)
	// b, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("error occurred: %v\n", err)
	}

	fmt.Fprintf(w, "%s", b)
}

/*
/ $ chronyc clients
Hostname                      NTP   Drop Int IntL Last     Cmd   Drop Int  Last
===============================================================================
testclient1                  187      0   9   -   128       0      0   -     -
testclient2                   78      0  10   -  1035       0      0   -     -
*/

func chrony_clients(w http.ResponseWriter, req *http.Request) {
	out := chrony_info("clients")

	scanner := bufio.NewScanner(strings.NewReader(string(out)))

	var clients []Client

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "Hostname") || strings.Contains(line, "==============================") {
			continue
		}

		kv := strings.Fields(line)

		if len(kv) == 10 {
			item := Client{
				Hostname:                       kv[0],
				NtpPacketsReceived:             kv[1],
				NtpPacketsDropped:              kv[2],
				NtpAverageInterval:             kv[3],
				NtpAverageIntervalAfterRRLimit: kv[4],
				NtpTimeSinceLastReceived:       kv[5],
				NtsKEReceived:                  kv[6],
				NtsKEDropped:                   kv[7],
				NtsKEAverageInterval:           kv[8],
				NtsKETimeSinceLastReceived:     kv[9],
			}
			clients = append(clients, item)
		}
	}

	result := &ClientsResult{
		Clients: clients,
	}

	b, err := json.Marshal(result)
	// b, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("error occurred: %v\n", err)
	}

	fmt.Fprintf(w, "%s", b)
}

/*
NTP packets received       : 330
NTP packets dropped        : 0
Command packets received   : 7087
Command packets dropped    : 0
Client log records dropped : 0
NTS-KE connections accepted: 0
NTS-KE connections dropped : 0
Authenticated NTP packets  : 0
Interleaved NTP packets    : 0
NTP timestamps held        : 0
NTP timestamp span         : 0
NTP daemon RX timestamps   : 0
NTP daemon TX timestamps   : 330
NTP kernel RX timestamps   : 330
NTP kernel TX timestamps   : 0
NTP hardware RX timestamps : 0
NTP hardware TX timestamps : 0
*/

func chrony_serverstats(w http.ResponseWriter, req *http.Request) {
	out := chrony_info("serverstats")

	scanner := bufio.NewScanner(strings.NewReader(string(out)))

	var result [17]KeyValue

	i := 0
	for scanner.Scan() {
		line := scanner.Text()
		kv := strings.Split(line, ": ")
		key := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ToLower(strings.TrimSpace(kv[0])), " ", "_"), ")", ""), "(", ""), "-", "")
		value := strings.TrimSpace(kv[1])

		pair := KeyValue{
			Key:   key,
			Value: value,
		}

		result[i] = pair

		i++

		if i > 16 {
			break
		}
	}

	data := &ServerStatsResult{
		Info: result,
	}

	// b, err := json.MarshalIndent(data, "", "  ")
	b, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("error occurred: %v\n", err)
	}

	fmt.Fprintf(w, "%s", b)
}

func chrony_config(w http.ResponseWriter, req *http.Request) {
	return_filename := "chrony.conf"
	filename := "/etc/chrony/" + return_filename

	switch req.Method {
	/*case "POST":
	fmt.Println("chrony config file upload")
	file, handler, err := req.FormFile("conf_filename")
	if err != nil {
		log.Println("Error retrieving the file:", err)
		return
	}
	defer file.Close()

	log.Printf("Uploaded File: %+v\n", handler.Filename)
	log.Printf("File Size:     %+v\n", handler.Size)
	log.Printf("MIME Header:   %+v\n", handler.Header)

	log.Println("File successfully uploaded.")

	/* bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		fmt.Println(err)
	}
	bodyString := string(bodyBytes)
	fmt.Printf("file content:\n=======================================\n%s\n=======================================\n", bodyString)
	*/
	case "GET":
		getFileContents(w, filename, return_filename, true)
	case "HEAD":
		getFileContents(w, filename, return_filename, false)
	default:
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}

	fmt.Fprintf(w, "success")
}

func chrony_logfile(w http.ResponseWriter, req *http.Request) {
	return_filename := "chrony.log"
	filename := "/var/log/chrony/" + return_filename

	switch req.Method {
	case "GET":
		result := getFileContents(w, filename, return_filename, true)

		if result.IsError {
			fmt.Println("GET LOGFILE ERROR: " + result.Text)
			http.Error(w, result.Text, result.HttpResponse)
		} else {
			fmt.Fprintf(w, "")
		}
	case "HEAD":
		result := getFileContents(w, filename, return_filename, false)

		if result.IsError {
			fmt.Println("HEAD LOGFILE ERROR: " + result.Text)
			http.Error(w, result.Text, result.HttpResponse)
		} else {
			fmt.Fprintf(w, "")
		}
	default:
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func getFileContents(w http.ResponseWriter, filename string, return_filename string, withContent bool) HttpServerStatus {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("failed to open file: %s\n", err)
		return HttpServerStatus{
			IsError:      true,
			Text:         `not found: ` + filename,
			HttpResponse: http.StatusNotFound,
			Err:          err,
		}
	}
	defer file.Close()

	w.Header().Add("Content-Type", "text/plain")
	w.Header().Add("Content-Disposition", `attachment; filename="`+return_filename+`"`)

	if withContent {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			// fmt.Println(line)
			fmt.Fprintln(w, line)
		}
		// Check for errors during the scan
		if err := scanner.Err(); err != nil {
			fmt.Printf("error reading file: %s", err)
			http.Error(w, "internal error reading file", http.StatusInternalServerError)
			return HttpServerStatus{
				IsError:      true,
				Text:         "internal error reading file",
				HttpResponse: http.StatusInternalServerError,
				Err:          err,
			}
		}
	}

	return HttpServerStatus{
		IsError:      false,
		Text:         "success",
		HttpResponse: http.StatusOK,
		Err:          nil,
	}
}

/*
func chrony_loglevel(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "POST":
		fmt.Println("chrony POST loglevel")

		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading body: %v\n", err)
			return
		}

		fmt.Println("HTTP Response Body:")
		fmt.Println(string(bodyBytes))
	case "GET":
		fmt.Println("chrony GET loglevel")
	default:
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	fmt.Fprintf(w, "success")
}
*/

func main() {
	fileServer := http.FileServer(http.Dir("/opt/www"))
	http.Handle("/", fileServer)
	http.Handle("/favicon.ico", http.NotFoundHandler())

	http.HandleFunc("/api/chrony/tracking", chrony_tracking)
	http.HandleFunc("/api/chrony/sourcestats", chrony_sourcestats)
	http.HandleFunc("/api/chrony/serverstats", chrony_serverstats)
	http.HandleFunc("/api/chrony/sources", chrony_sources)
	http.HandleFunc("/api/chrony/clients", chrony_clients)
	http.HandleFunc("/api/chrony/logfile", chrony_logfile)
	http.HandleFunc("/api/chrony/configfile", chrony_config)
	// http.HandleFunc("/api/chrony/loglevel", chrony_loglevel)

	fmt.Printf("Starting server at port 80\n")
	if err := http.ListenAndServe(":80", nil); err != nil {
		log.Fatal(err)
	}
}
