package goawspricing

import (
	"io/ioutil"
	"strconv"

	"github.com/bitly/go-simplejson"
	"github.com/mob-manzoku/go-jsonp"
	"gopkg.in/yaml.v2"
)

type InstanceTypes map[string]*Size

type Size struct {
	Size           *string  `json:"size"`
	VCPU           *int     `json:"vCPU"`
	ECU            *string  `json:"ECU"`
	MemoryGiB      *float64 `json:"memoryGiB"`
	StorageGB      *string  `json:"storageGB"`
	System         *string  `json:"system"`
	PiopsOptimized *bool    `json:"piopsOptimized"`
	Price          *float64 `json:"price"`
	Network        *string  `json:"network"`
}

var awsPricingURLs = map[string]string{
	"ec2":         "http://a0.awsstatic.com/pricing/1/ec2/linux-od.min.js",
	"rds":         "http://a0.awsstatic.com/pricing/1/rds/mysql/pricing-standard-deployments.min.js",
	"elasticache": "http://a0.awsstatic.com/pricing/1/elasticache/pricing-standard-deployments-elasticache.min.js",
}

func GetEC2Pricing(region string) InstanceTypes {

	service := "ec2"
	ret := InstanceTypes{}

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

		types := regions.GetIndex(i).Get("instanceTypes")

		for j := range types.MustArray() {

			sizes := types.GetIndex(j).Get("sizes")
			for k := range sizes.MustArray() {
				s := sizes.GetIndex(k)

				msize := s.Get("size").MustString()
				mvCPU, _ := strconv.Atoi(s.Get("vCPU").MustString())
				mECU := s.Get("ECU").MustString()
				mmem, _ := strconv.ParseFloat(s.Get("memoryGiB").MustString(), 64)
				msto := s.Get("storageGB").MustString()
				msys := s.Get("valueColumns").GetIndex(0).Get("name").MustString()
				mprice, _ := strconv.ParseFloat(s.Get("valueColumns").GetIndex(0).Get("prices").Get("USD").MustString(), 64)

				obj := &Size{
					Size:      &msize,
					VCPU:      &mvCPU,
					ECU:       &mECU,
					MemoryGiB: &mmem,
					StorageGB: &msto,
					System:    &msys,
					Price:     &mprice,
				}

				ret[*obj.Size] = obj

			}
		}

	}
	return ret

}

func GetElasticachePricing(region string) InstanceTypes {

	service := "elasticache"
	specpath := "spec/elasticache.yml"
	ret := InstanceTypes{}

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
				msys := "Redis"
				obj := &Size{
					Size:   &msize,
					Price:  &mprice,
					System: &msys,
				}

				ret[*obj.Size] = obj
			}

		}
	}

	buf, err := ioutil.ReadFile(specpath)
	if err != nil {
		return ret
	}

	m := make(map[interface{}]interface{})
	err = yaml.Unmarshal(buf, &m)

	for _, v := range m[service].([]interface{}) {

		msize := v.(map[interface{}]interface{})["size"].(string)
		mvCPU := v.(map[interface{}]interface{})["vCPU"].(int)
		mmem := v.(map[interface{}]interface{})["memoryGiB"].(float64)
		mnet := v.(map[interface{}]interface{})["network"].(string)

		if _, ok := ret[msize]; ok {
			ret[msize].VCPU = &mvCPU
			ret[msize].MemoryGiB = &mmem
			ret[msize].Network = &mnet
		}

	}

	return ret
}

func GetRDSPricing(region string) InstanceTypes {

	service := "rds"
	specpath := "spec/rds.yml"
	ret := InstanceTypes{}

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
				msys := "MySQL"
				obj := &Size{
					Size:   &msize,
					Price:  &mprice,
					System: &msys,
				}

				ret[*obj.Size] = obj
			}

		}
	}

	buf, err := ioutil.ReadFile(specpath)
	if err != nil {
		return ret
	}

	m := make(map[interface{}]interface{})
	err = yaml.Unmarshal(buf, &m)

	for _, v := range m[service].([]interface{}) {

		msize := v.(map[interface{}]interface{})["size"].(string)
		mvCPU := v.(map[interface{}]interface{})["vCPU"].(int)
		mpiop := v.(map[interface{}]interface{})["piopsOptimized"].(bool)
		mmem := v.(map[interface{}]interface{})["memoryGiB"].(float64)
		mnet := v.(map[interface{}]interface{})["network"].(string)

		if _, ok := ret[msize]; ok {
			ret[msize].VCPU = &mvCPU
			ret[msize].MemoryGiB = &mmem
			ret[msize].PiopsOptimized = &mpiop
			ret[msize].Network = &mnet
		}

	}

	return ret
}
