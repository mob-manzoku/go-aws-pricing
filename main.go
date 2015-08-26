package main

import (
	"fmt"

	"github.com/bitly/go-simplejson"
	"github.com/mob-manzoku/go-jsonp"
)

type instanceTypes map[string]size

type size struct {
	size           string
	vCPU           string
	ECU            string
	memoryGiB      string
	storageGB      string
	system         string
	piopsOptimized string
	price          string
}

var awsPricingURLs = map[string]string{
	"ec2":         "http://a0.awsstatic.com/pricing/1/ec2/linux-od.min.js",
	"rds":         "http://a0.awsstatic.com/pricing/1/rds/mysql/pricing-standard-deployments.min.js",
	"elasticache": "http://a0.awsstatic.com/pricing/1/elasticache/pricing-standard-deployments-elasticache.min.js",
}

func main() {

	region := "apac-tokyo"

	str, _ := gojsonp.GetJSONFromURL(awsPricingURLs["rds"])

	raw, err := simplejson.NewJson([]byte(str))
	if err != nil {
		panic(err)
	}

	regions := raw.Get("config").Get("regions")

	for i := range regions.MustArray() {

		if regions.GetIndex(i).Get("region").MustString() != region {
			continue
		}

		types := regions.GetIndex(i).Get("types")

		for j := range types.MustArray() {

			tiers := types.GetIndex(j).Get("tiers")

			for k := range tiers.MustArray() {
				fmt.Println(tiers.GetIndex(k).Get("name"))
				fmt.Println(tiers.GetIndex(k).Get("prices").Get("USD"))
			}

		}
	}

	// path := "static/rds_spec.yml"
	// buf, err := ioutil.ReadFile(path)
	// if err != nil {
	// 	panic(err)
	// }

	// m := make(map[interface{}]interface{})
	// err = yaml.Unmarshal(buf, &m)

	// ret := instanceTypes{}

	// for _, v := range m["rds"].([]interface{}) {

	// 	obj := size{
	// 		size: v.(map[interface{}]interface{})["size"].(string),
	// 	}

	// 	ret[obj.size] = obj
	// }

}
