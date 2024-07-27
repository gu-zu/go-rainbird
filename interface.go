package rainbird

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/bits"
	"time"
)

// == DATA TYPES ==

type irrigationInterval int

const (
	Icustom irrigationInterval = iota
	Iodd
	Ieven
	Icyclic
)

var IntervalName = map[irrigationInterval]string{Icustom: "custom", Iodd: "odd", Ieven: "even", Icyclic: "cyclic"}

type Schedule struct {
	Duration   time.Duration // duration this zone will be turned on
	Time       []time.Time   // max 6, only hour and minute are used, year and date are always 0
	Interval   irrigationInterval
	customDays byte
}

func (sched *Schedule) IsActive(day int) bool {
	if day == 6 { // SUNDAY IS NOT THE FIRST DAY OF THE WEEK
		day = -1
	}
	if day < 0 || day > 6 {
		return false // not a valid day
	}
	return (sched.customDays & (1 << (day + 1))) != 0
}

func (sched *Schedule) SetActive(day int) {
	if day == 6 { // SUNDAY IS NOT THE FIRST DAY OF THE WEEK
		day = -1
	}
	if day < 0 || day > 6 {
		return // not a valid day
	}
	sched.customDays = sched.customDays | (1 << (day + 1))
}

func (sched *Schedule) String() string {
	return fmt.Sprintf("Schedule: ontime: %s, at: %s, schedule: %s [mon:%t tue:%t wed:%t thu:%t fri:%t sat:%t sun:%t]", sched.Duration.String(), sched.Time, IntervalName[sched.Interval], sched.IsActive(0), sched.IsActive(1), sched.IsActive(2), sched.IsActive(3), sched.IsActive(4), sched.IsActive(5), sched.IsActive(6))
}

// == DEVICE CMD FUNCTIONS ==

func (rb *device) GetModelandVersion() (string, error) {
	res, err := rb.message("02", "82")
	if err != nil {
		return "", err
	}
	model := map[string]string{"0003": "ESP-RZXe", "0007": "ESP-Me", "0006": "ST8x-WiFi", "0005": "ESP-TM2", "0008": "St8x-WiFi2", "0009": "ESP-ME3", "0010": "ESP=Me2", "000a": "ESP-TM2", "010a": "ESP-TM2", "0099": "TBOS-BT", "0107": "ESP-Me", "0103": "ESP-RZXe2", "0812": "ARC8"}[hex.EncodeToString(res[1:3])]
	return fmt.Sprintf("%s, %d.%d", model, res[3], res[4]), nil
}

// returns the number of active zone, if none active returns 0
func (rb *device) GetCurrentState() (int, error) {
	res, err := rb.message("3F00", "BF")
	if err != nil {
		return 0, err
	}
	zone := bits.TrailingZeros8(res[3]) + 1
	if zone == 9 {
		zone = 0
	}
	return zone, nil
}

func (rb *device) GetSchedule(zone int) (*Schedule, error) {
	res, err := rb.message("20000"+fmt.Sprint(zone), "A0")
	if len(res) != 14 {
		return new(Schedule), fmt.Errorf("invalid rainbird response: %v", res)
	}
	if err != nil {
		return new(Schedule), err
	}
	if int(res[2]) != zone {
		return new(Schedule), fmt.Errorf("invalid rainbird response zone: %d->%d", zone, res[2])
	}
	sched := &Schedule{time.Duration(res[3]) * time.Minute, []time.Time{}, irrigationInterval(res[10]), res[11]}
	for i := 4; i < 10; i++ {
		if res[i] != 144 {
			t := time.Date(0, 1, 1, int(math.Floor(float64(res[i])/6)), int(res[i]%6)*10, 0, 0, time.Local)
			sched.Time = append(sched.Time, t)
		}
	}
	return sched, nil
}

func (rb *device) SetSchedule(zone int, Schedule *Schedule) error {
	msg := make([]byte, 12)
	msg[0] = byte(zone)
	msg[1] = byte(Schedule.Duration.Minutes())
	for i := 0; i < 6; i++ {
		if i < len(Schedule.Time) {
			msg[i+2] = byte(Schedule.Time[i].Hour()*6 + int(Schedule.Time[i].Minute()/10))
		} else {
			msg[i+2] = 144
		}
	}
	msg[8] = byte(Schedule.Interval)
	msg[9] = Schedule.customDays
	msg[10] = 2 // TODO - fix this byte, seems to be soil type
	_, err := rb.message("2100"+hex.EncodeToString(msg), "01")
	if err != nil {
		return err
	}
	return nil
}

// TODO - find out if 100 min irrigation time is max of app or max of controller
func (rb *device) RunManual(zone int, minutes int) error {
	_, err := rb.message(fmt.Sprintf("39000%d%s", zone, hex.EncodeToString([]byte{byte(minutes)})), "01")
	if err != nil {
		return err
	}
	return nil
}

func (rb *device) StopManual(zone int) error {
	_, err := rb.message("40", "01")
	if err != nil {
		return err
	}
	return nil
}

func (rb *device) GetTime() (time.Time, error) {
	res, err := rb.message("10", "90")
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(0, 1, 1, int(res[1]), int(res[2]), int(res[3]), 0, time.Local), err
}
func (rb *device) SetTime(t time.Time) error {
	data := []byte{byte(t.Hour()), byte(t.Minute()), byte(t.Second())}
	_, err := rb.message("11"+hex.EncodeToString(data), "01")
	if err != nil {
		return err
	}
	return nil
}

func (rb *device) GetDate() (time.Time, error) {
	res, err := rb.message("12", "92")
	if err != nil {
		return time.Time{}, err
	}
	str := hex.EncodeToString(res)
	day, _ := hex.DecodeString(str[2:4])
	mth, _ := hex.DecodeString("0" + str[4:5])
	yr, _ := hex.DecodeString("0" + str[5:8])
	return time.Date(int(yr[0])*256+int(yr[1]), time.Month(mth[0]), int(day[0]), 0, 0, 0, 0, time.Local), nil
}

func (rb *device) SetDate(t time.Time) error {
	_, err := rb.message("13"+hex.EncodeToString([]byte{byte(t.Day())})+hex.EncodeToString([]byte{byte(t.Month())})[1:2]+hex.EncodeToString([]byte{byte(t.Year() / 256), byte(t.Year() % 256)})[1:4], "01")
	if err != nil {
		return err
	}
	return nil
}

// Returns raindelay in days
func (rb *device) GetRainDelay() (int, error) {
	res, err := rb.message("36", "B6")
	if err != nil {
		return 0, err
	}
	if len(res) != 3 {
		return 0, fmt.Errorf("invalid rainbird response: %v", res)
	}
	fmt.Println(res)
	return int(res[1])*256 + int(res[2]), nil
}
func (rb *device) SetRainDelay(days byte) error {
	_, err := rb.message("3700"+hex.EncodeToString([]byte{days}), "01")
	if err != nil {
		return err
	}
	return nil
}

func (rb *device) GetWifi() (*wifiResult, error) {
	return rb.methodmsg("getWifiParams")
}

/*TODO

02 -> done
03 -> todo
04 -> todo
05 -> todo
0B -> NOT SUPPORTED BY RZXe
10 -> done
11 -> done
12 -> done
13 -> done
20 -> done
21 -> done
22 -> NOT SUPPORTED BY RZXe
30 -> todo
32 -> NOT SUPPORTED BY RZXe
36 -> done
37 -> done
3E -> todo
3F -> done
38 -> NOT SUPPORTED BY RZXe
39 -> done
3A -> todo
3B -> todo
40 -> done
42 -> NOT SUPPORTED BY RZXe
48 -> todo
49 -> NOT SUPPORTED BY RZXe
4A -> NOT SUPPORTED BY RZXe
4B -> NOT SUPPORTED BY RZXe
4C -> NOT SUPPORTED BY RZXe


*/
