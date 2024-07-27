package rainbird

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
)

func TestPackageMsg(t *testing.T) {
	result := packageMsg(5, "somedata")
	var f interface{}
	err := json.Unmarshal([]byte(result), &f)
	if err != nil {
		t.Error("error converting to json", err)
	}
	t.Log("packagedata result:", result, f)
}

func TestCrypt(t *testing.T) {
	result := encrypt(`{"id": 5, "method": "tunnelSip", "params": {"length": "4", "data": "somedata"}, "jsonrpc": "2.0"}`, "*******")
	hexr := make([]byte, hex.EncodedLen(len(result)))
	hex.Encode(hexr, result)
	t.Log("encrypt result:", string(hexr))
	decrypted := decrypt(result, "*******")
	t.Log("decrypt result:", string(decrypted))
	if string(decrypted) != `{"id": 5, "method": "tunnelSip", "params": {"length": "4", "data": "somedata"}, "jsonrpc": "2.0"}` {
		t.Error("incorrectly decrypted")
	}
	//{"id": 5, "method": "tunnelSip", "params": {"length": "4", "data": "somedata"}, "jsonrpc": "2.0"}
	//first encrypt data, then check if decrypt results in original data
}

// test sending a basic command(request for model and version)
func TestCmd(t *testing.T) {
	rb := Get("10.0.0.55", "*******")
	res, err := rb.message("02", "82")
	if err != nil {
		t.Error(err)
	}
	t.Log("02 Message response/err", res, err)
	rb = Get("10.0.0.55", "ehNekv9")
	res, err = rb.message("02", "82")
	if err == nil {
		t.Error("did not fail with incorrect credentials", res)
	}
	t.Log("02 Message response/err(unauth)", res, err)
}

// Get settings of rb and store in a file
func TestRbState(t *testing.T) {
	rb := Get("10.0.0.55", "*******")
	cmds := []string{"02", "03", "04", "05", "0B", "10", "11", "12", "13", "20", "21", "22", "30", "32", "36", "37", "3E", "3F", "38", "39", "3A", "3B", "40", "42", "48", "49", "4A", "4B", "4C"}
	file, err := os.Create("rb" + strings.ReplaceAll(rb.ip, ".", "_") + ".state")
	if err != nil {
		t.Fatal("error opening results file", err)
	}
	for i := 0; i < len(cmds); i++ {
		res, err := rb.message(cmds[i], "")
		if err != nil {
			res = []byte("Error: " + err.Error())
		}
		file.WriteString(cmds[i] + " -> " + string(res) + "\n")
	}
}

// test all implemented rb functions
// ! This test will run zone 2 for 20 seconds, set the time/date to the machine's current time and add 1 day to raindelay
func TestAllCmdFuncs(t *testing.T) {
	rb := Get("10.0.0.55", "*******")
	str, err := rb.GetModelandVersion()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("modelandver", str)

	// == schedule test ==
	schedule, err := rb.GetSchedule(1)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(schedule)
	schedule.Time = append(schedule.Time, schedule.Time[len(schedule.Time)-1].Add(time.Hour))
	err = rb.SetSchedule(1, schedule)
	if err != nil {
		t.Fatal(err)
	}
	scheduleNew, err := rb.GetSchedule(1)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("schedule new", scheduleNew)
	if !scheduleNew.Time[len(scheduleNew.Time)-1].Equal(schedule.Time[len(schedule.Time)-1]) {
		t.Fatal("time did not update ")
	}
	schedule.Time = schedule.Time[:len(schedule.Time)-1]
	err = rb.SetSchedule(1, schedule)
	if err != nil {
		t.Fatal(err)
	}

	err = rb.RunManual(2, 4) // zone 2, 4 min
	if err != nil {
		t.Fatal(err)
	}
	num, err := rb.GetCurrentState()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("curstate", num)
	time.Sleep(time.Second * 20)
	err = rb.StopManual(2)
	if err != nil {
		t.Fatal(err)
	}
	err = rb.SetTime(time.Now())
	if err != nil {
		t.Fatal(err)
	}
	rbtime, err := rb.GetTime()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("time", rbtime)

	err = rb.SetDate(time.Now())
	if err != nil {
		t.Fatal(err)
	}
	date, err := rb.GetDate()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("date", date)

	wifi, err := rb.GetWifi()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v\n", wifi)

	days, err := rb.GetRainDelay()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("raindelaydays", days)
	err = rb.SetRainDelay(byte(days + 1))
	if err != nil {
		t.Fatal(err)
	}
	days2, err := rb.GetRainDelay()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("raindelaydays", days2)
	if days2-1 != days {
		t.Fatal("raindelay did not update")
	}
}
