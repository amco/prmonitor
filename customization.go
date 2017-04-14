package prmonitor

func getCustomizations() (Customization) {
	return Customization{
		passiveColor: "#00cc66",
		warningColor: "#ffff00",
		alertColor:   "#cc0000",
		passiveTime:  24.0,
		warningTime:  48,
	}
}

type Customization struct {
	passiveColor	string		// #00cc66"
	warningColor	string		// #ffff00
	alertColor	string		// #cc0000
	passiveTime	float64		// 24.0
	warningTime	float64		// 48
}
