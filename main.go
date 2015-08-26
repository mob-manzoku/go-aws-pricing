package main

import (
	"io/ioutil"
	"strconv"

	"gopkg.in/yaml.v2"

	"github.com/bitly/go-simplejson"
	"github.com/k0kubun/pp"
	"github.com/mob-manzoku/go-jsonp"
)

type instanceTypes map[string]*size

type size struct {
	size           *string
	vCPU           *int
	ECU            *float64
	memoryGiB      *float64
	storageGB      *float64
	system         *string
	piopsOptimized *bool
	price          *float64
	network        *string
}

var awsPricingURLs = map[string]string{
	"ec2":         "http://a0.awsstatic.com/pricing/1/ec2/linux-od.min.js",
	"rds":         "http://a0.awsstatic.com/pricing/1/rds/mysql/pricing-standard-deployments.min.js",
	"elasticache": "http://a0.awsstatic.com/pricing/1/elasticache/pricing-standard-deployments-elasticache.min.js",
}

func main() {

	region := "apac-tokyo"
	service := "rds"
	ret := instanceTypes{}

	str, _ := gojsonp.GetJSONFromURL(awsPricingURLs[service])

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

				msize := tiers.GetIndex(k).Get("name").MustString()
				mprice, _ := strconv.ParseFloat(tiers.GetIndex(k).Get("prices").Get("USD").MustString(), 64)
				obj := &size{
					size:  &msize,
					price: &mprice,
				}

				ret[*obj.size] = obj
			}

		}
	}

	path := "spec/rds.yml"
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	m := make(map[interface{}]interface{})
	err = yaml.Unmarshal(buf, &m)

	for _, v := range m[service].([]interface{}) {

		msize := v.(map[interface{}]interface{})["size"].(string)
		mvCPU := v.(map[interface{}]interface{})["vCPU"].(int)
		mmem := v.(map[interface{}]interface{})["memoryGiB"].(float64)
		mpiop := v.(map[interface{}]interface{})["piopsOptimized"].(bool)
		mnet := v.(map[interface{}]interface{})["network"].(string)

		if _, ok := ret[msize]; ok {
			ret[msize].vCPU = &mvCPU
			ret[msize].memoryGiB = &mmem
			ret[msize].piopsOptimized = &mpiop
			ret[msize].network = &mnet
		}

	}

	for k, v := range ret {
		pp.Print(k, *v)
	}

}
