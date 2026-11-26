package models

type ServerProfile struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	URL        string `json:"url"`
	Username   string `json:"username,omitempty"`
	LastSyncAt string `json:"lastSyncAt,omitempty"`
}

type HealthInfo struct {
	Status     string `json:"status"`
	Service    string `json:"service"`
	Version    string `json:"version"`
	ServerName string `json:"server_name"`
	ServerTime string `json:"server_time"`
}

type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FullName  string `json:"full_name"`
	Role      string `json:"role"`
	Active    bool   `json:"active"`
	CreatedAt string `json:"created_at"`
}

type PhaseInfo struct {
	Code        string `json:"code"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Color       string `json:"color"`
	ProgressPct int    `json:"progress_pct"`
}

type Alert struct {
	Code     string `json:"code"`
	Severity string `json:"severity"`
	Title    string `json:"title"`
	Message  string `json:"message"`
}

type Evaluation struct {
	Phase                  PhaseInfo `json:"phase"`
	RefTemperatureC        *float64  `json:"ref_temperature_c"`
	HasReachedThermophilic bool      `json:"has_reached_thermophilic"`
	PeakTemperatureC       *float64  `json:"peak_temperature_c"`
	DaysElapsed            int       `json:"days_elapsed"`
	Alerts                 []Alert   `json:"alerts"`
}

type Reading struct {
	ID           int64    `json:"id"`
	PileID       int64    `json:"pile_id"`
	TemperatureC float64  `json:"temperature_c"`
	HumidityPct  float64  `json:"humidity_pct"`
	PH           *float64 `json:"ph"`
	Phase        string   `json:"phase"`
	PowerSource  string   `json:"power_source"`
	BatteryV     *float64 `json:"battery_v"`
	RSSI         *int     `json:"rssi"`
	Seq          *int     `json:"seq"`
	RecordedAt   string   `json:"recorded_at"`
}

type Pile struct {
	ID            int64      `json:"id"`
	Name          string     `json:"name"`
	Location      string     `json:"location"`
	Substrate     string     `json:"substrate"`
	NodeID        string     `json:"node_id"`
	HasPHSensor   bool       `json:"has_ph_sensor"`
	Status        string     `json:"status"`
	Notes         string     `json:"notes"`
	StartedAt     string     `json:"started_at"`
	CreatedAt     string     `json:"created_at"`
	UpdatedAt     string     `json:"updated_at"`
	LatestReading *Reading   `json:"latest_reading"`
	Evaluation    Evaluation `json:"evaluation"`
}

type Device struct {
	ID         int64   `json:"id"`
	Name       string  `json:"name"`
	KeyPrefix  string  `json:"key_prefix"`
	Active     bool    `json:"active"`
	LastSeenAt *string `json:"last_seen_at"`
	CreatedAt  string  `json:"created_at"`
}

type PileInput struct {
	Name        string `json:"name"`
	Location    string `json:"location"`
	Substrate   string `json:"substrate"`
	NodeID      string `json:"node_id"`
	HasPHSensor bool   `json:"has_ph_sensor"`
	Notes       string `json:"notes"`
	Status      string `json:"status,omitempty"`
	StartedAt   string `json:"started_at,omitempty"`
}

type ManualReadingInput struct {
	TemperatureC float64  `json:"temperature_c"`
	HumidityPct  float64  `json:"humidity_pct"`
	PH           *float64 `json:"ph"`
}

type CreateUserInput struct {
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type LoginResult struct {
	User       User   `json:"user"`
	LastSyncAt string `json:"lastSyncAt"`
}

type SessionRestoreResult struct {
	User       User    `json:"user"`
	Offline    bool    `json:"offline"`
	LastSyncAt *string `json:"lastSyncAt"`
}

type PilesResponse struct {
	Piles      []Pile  `json:"piles"`
	Offline    bool    `json:"offline"`
	LastSyncAt *string `json:"lastSyncAt"`
}

type ReadingsResponse struct {
	Readings []Reading `json:"readings"`
	Offline  bool      `json:"offline"`
}

type CreateDeviceResult struct {
	Device Device `json:"device"`
	APIKey string `json:"api_key"`
}
