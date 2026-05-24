package namelookup

import (
  "encoding/json"
  "os"
  "strings"
)

type DemoPerson struct {
  FullName string            `json:"full_name"`
  Address  string            `json:"address"`
  Phone    string            `json:"phone"`
  Extra    map[string]string `json:"extra"`
}

func lookupDemoPerson(fullname string) (DemoPerson, bool) {
  f, err := os.Open("data/demo_people.json")
  if err != nil {
    return DemoPerson{}, false
  }
  defer f.Close()
  var people []DemoPerson
  if err := json.NewDecoder(f).Decode(&people); err != nil {
    return DemoPerson{}, false
  }
  for _, p := range people {
    if strings.EqualFold(strings.TrimSpace(p.FullName), strings.TrimSpace(fullname)) {
      return p, true
    }
  }
  return DemoPerson{}, false
}
