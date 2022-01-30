package procon

type Input struct {
	LStick    LStick `json:"L_STICK"`
	RStick    RStick `json:"R_STICK"`
	DpadUp    bool   `json:"DPAD_UP"`
	DpadLeft  bool   `json:"DPAD_LEFT"`
	DpadRight bool   `json:"DPAD_RIGHT"`
	DpadDown  bool   `json:"DPAD_DOWN"`
	L         bool   `json:"L"`
	Zl        bool   `json:"ZL"`
	R         bool   `json:"R"`
	Zr        bool   `json:"ZR"`
	JclSr     bool   `json:"JCL_SR"`
	JclSl     bool   `json:"JCL_SL"`
	JcrSr     bool   `json:"JCR_SR"`
	JcrSl     bool   `json:"JCR_SL"`
	Plus      bool   `json:"PLUS"`
	Minus     bool   `json:"MINUS"`
	Home      bool   `json:"HOME"`
	Capture   bool   `json:"CAPTURE"`
	Y         bool   `json:"Y"`
	X         bool   `json:"X"`
	B         bool   `json:"B"`
	A         bool   `json:"A"`
}

type LStick struct {
	Pressed bool `json:"PRESSED"`
	XValue  int  `json:"X_VALUE"`
	YValue  int  `json:"Y_VALUE"`
	LsUp    bool `json:"LS_UP"`
	LsLeft  bool `json:"LS_LEFT"`
	LsRight bool `json:"LS_RIGHT"`
	LsDown  bool `json:"LS_DOWN"`
}

type RStick struct {
	Pressed bool `json:"PRESSED"`
	XValue  int  `json:"X_VALUE"`
	YValue  int  `json:"Y_VALUE"`
	RsUp    bool `json:"RS_UP"`
	RsLeft  bool `json:"RS_LEFT"`
	RsRight bool `json:"RS_RIGHT"`
	RsDown  bool `json:"RS_DOWN"`
}
