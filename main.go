package main

import (
  ui "github.com/gizak/termui"
  "time"
  "github.com/common-nighthawk/go-figure"
  "encoding/json"
  "net/http"
)

// Custom type for JSON decoding
type jsonSchedules struct {
  Response struct {
    Schedules[] struct {
      Message string
    }
  }
}

type jsonTraffic struct {
  Response struct {
    Message string
  }
}

type jsonToday struct {
  Sys struct {
    Sunrise int64
    Sunset int64
  }
}

type jsonWeather struct {
  List []struct {
    Dt int64
    Main struct {
      Temp float64
    }
    Clouds struct {
      All int64
    }
    Rain struct {
      Total float64 `json:"3h"`
    }
    Wind struct {
      Speed float64
    }
  }
}

// Globals

var termHeight int = 0
var termWidth int = 0
var accentColor ui.Attribute = ui.ColorCyan
var types = []string{ "rer", "bus" }
var timeClockFormat string = "15:04:05"
var sunTimeFormat string = "15h04"
var dateFormat string = "Monday 2 January 2006"
var urlSchedulesRer string = "http://api-ratp.pierre-grimaud.fr/v2/rers/a/stations/10?destination=1"
var urlSchedulesBus string = "https://api-ratp.pierre-grimaud.fr/v2/bus/124/stations/1596?destination=108"
var urlTrafficRer string = "https://api-ratp.pierre-grimaud.fr/v2/traffic/rers/A"
var urlWeather string = "http://api.openweathermap.org/data/2.5/forecast?id=6452019&mode=json&appid=6e2218dcec22c786e4a039dfe3bfae98&lang=fr&units=metric"
var urlToday string = "http://api.openweathermap.org/data/2.5/weather?id=6452019&appid=6e2218dcec22c786e4a039dfe3bfae98"


// Funcs
func newMyList(title string, colorFG ui.Attribute, border bool) * ui.List {
  tmpList := ui.NewList()
  tmpList.BorderLabel = title
  tmpList.Border = border
  tmpList.BorderLabelFg = accentColor
  tmpList.ItemFgColor = colorFG
  return tmpList
}

func newMyGraph(title string, barcolor ui.Attribute) * ui.BarChart {
  gTmp := ui.NewBarChart()
  gTmp.Height = termHeight / 4
  gTmp.BorderLabel = title
  gTmp.BorderLabelFg = accentColor
  gTmp.BarColor = barcolor
  gTmp.TextColor = barcolor
  gTmp.NumColor =  ui.ColorWhite
  gTmp.BarGap = 1
  gTmp.BarWidth = 5
  return gTmp
}

func runNowAndEvery(seconds int, function func()) {
  go func() {
    time.Sleep(1 * time.Second)
    function()
    ticker := time.NewTicker(time.Duration(seconds) * time.Second)
    for range ticker.C {
      function()
    }
  }()
}

func centerList(myList * ui.List, verticaly bool) {
  if verticaly {
    myList.PaddingTop = (myList.Height - len(myList.Items)) / 2
  }
  maxChars := 0
  for _, value := range myList.Items {
    l := len(value)
    if l > maxChars {
      maxChars = l
    }
  }
  myList.PaddingLeft = (myList.Width - maxChars) / 2
}

func centerPar(myPar * ui.Par) {
  l := len(myPar.Text)
  if l < myPar.Width {
    myPar.PaddingTop = 1
    myPar.PaddingLeft = (myPar.Width - l) / 2
  } else {
    myPar.PaddingTop = 0
    myPar.PaddingLeft = 0
  }
}

func getJson(url string, target interface {}) error {
  r, err := http.Get(url)
  if err != nil {
    return err
  }
  defer r.Body.Close()
  return json.NewDecoder(r.Body).Decode(target)
}

func main() {

  // Init Ui

  err := ui.Init()
  if err != nil {
    panic(err)
  }
  defer ui.Close()


  // Init vars

  var urlSchedules = make(map[string] string)
  urlSchedules["rer"] = urlSchedulesRer
  urlSchedules["bus"] = urlSchedulesBus

  termHeight = ui.TermHeight()
  termWidth = ui.TermWidth()

  // Bus & RER Schedules

  var ls = make(map[string] * ui.List)
  ls["bus"] = newMyList("Bus -> VDF", ui.ColorYellow, true)
  ls["rer"] = newMyList("RER -> Paris", ui.ColorRed, true)

  trafficRER := ui.NewPar("")
  trafficRER.BorderLabel = "Traffic RER"
  trafficRER.TextFgColor = ui.ColorRed
  trafficRER.BorderLabelFg = accentColor

  // Today

  lsToday := newMyList("Aujourd'hui", ui.ColorGreen, true)

  // Clock

  lsTime := newMyList("", ui.ColorWhite, false)

  // Weather Graphs

  gTemp := newMyGraph("Températures (°C)", ui.ColorRed)
  gCloud := newMyGraph("Couverture nuageuse (%)", ui.ColorYellow)
  gWind := newMyGraph("Vent (m/s)", ui.ColorCyan)
  gRain := newMyGraph("Pluie (mm x10)", ui.ColorBlue)
  gTemp.Height = termHeight - gCloud.Height - gWind.Height - gRain.Height


  // Updates

  updateClock := func () {
      t := time.Now()
      str := t.Format(timeClockFormat)
      myFigure := figure.NewFigure(str, "ogre", false)
      lsTime.Items = myFigure.Slicify()
      centerList(lsTime, true)
      ui.Render(ui.Body)
  }
  runNowAndEvery(1, updateClock)


  updateSchedulesAndTraffic := func () {
    // RER & BUS
    for _, value := range types {
      s := new(jsonSchedules)
      getJson(urlSchedules[value], s)
      ss := make([]string, len(s.Response.Schedules))
      for key,
      _ := range s.Response.Schedules {
        ss[key] = s.Response.Schedules[key].Message
      }
      ls[value].Items = ss
      centerList(ls[value], false)
    }
    // Traffic
    s := new(jsonTraffic)
    getJson(urlTrafficRer, s)
    trafficRER.Text = s.Response.Message
    centerPar(trafficRER)
    // Finally
    ui.Render(ui.Body)
  }
  runNowAndEvery(15, updateSchedulesAndTraffic)

  updateWeatherAndToday := func () {
    now := time.Now()
    // Weather
    s := new(jsonWeather)
    getJson(urlWeather, s)
    cDay := now.Day()
    temp := make([]int, len(s.List))
    cloud := make([]int, len(s.List))
    rain := make([]int, len(s.List))
    wind := make([]int, len(s.List))
    lb := make([]string, len(s.List))
    for key,
    _ := range s.List {
      tm := time.Unix(s.List[key].Dt, 0)
      var label string
      switch tm.Day() {
        case cDay:
          label = "A "
        case cDay + 1:
          label = "D "
        case cDay + 2:
          label = "P "
        default:
          label = "O "
      }
      lb[key] = label + tm.Format("15h")
      temp[key] = int(s.List[key].Main.Temp)
      cloud[key] = int(s.List[key].Clouds.All)
      rain[key] = int(s.List[key].Rain.Total * 10)
      wind[key] = int(s.List[key].Wind.Speed)
    }
    gTemp.Data = temp
    gCloud.Data = cloud
    gRain.Data = rain
    gWind.Data = wind
    gTemp.DataLabels, gCloud.DataLabels = lb, lb
    gRain.DataLabels, gWind.DataLabels = lb, lb
    // Today
    s2 := new(jsonToday)
    getJson(urlToday, s2)
    ss := time.Unix(s2.Sys.Sunset, 0)
    sr := time.Unix(s2.Sys.Sunrise, 0)
    lsToday.Items = []string{
      now.Format(dateFormat),
      "Lever du soleil : " + sr.Format(sunTimeFormat),
      "Coucher du soleil : " + ss.Format(sunTimeFormat),
    }
    centerList(lsToday, false)
    ui.Render(ui.Body)
  }
  runNowAndEvery(60 * 10, updateWeatherAndToday)


  // Layout

  ui.Body.AddRows(
    ui.NewRow(
      ui.NewCol(6, 0, lsTime, lsToday, trafficRER, ls["rer"], ls["bus"]),
      ui.NewCol(6, 0, gTemp, gCloud, gWind, gRain),
    ),
  )
  ui.Body.Align()

  // Ajust Layout

  heightSchedules := 8
  ls["bus"].Height = heightSchedules
  ls["rer"].Height = heightSchedules
  ls["bus"].Y = termHeight - heightSchedules
  ls["rer"].Y = termHeight - heightSchedules

  widthSchedules := termWidth / 4
  ls["rer"].Width  = widthSchedules
  ls["bus"].Width  = widthSchedules
  ls["bus"].X = widthSchedules

  heightTrafficRer := 5
  trafficRER.Height = heightTrafficRer
  trafficRER.Y = ls["rer"].Y - heightTrafficRer

  heightToday := 5
  lsToday.Height = heightToday
  lsToday.Y = trafficRER.Y - heightToday

  lsTime.Height = termHeight - heightSchedules - heightTrafficRer - heightToday

  // Render

  ui.Render(ui.Body)

  // Events

  ui.Handle("/sys/kbd/q", func(ui.Event) {
    ui.StopLoop()
  })

  // Loop !
  ui.Loop()

}
