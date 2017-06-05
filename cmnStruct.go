package main

func createMapUser() map[string] interface{} {
	retMap:=map[string] interface{}{
		"uid":1,
		"coin":30,
		"energy":3 ,
		"max_energy":15 ,
		"last_recover" :0,
		"day_recover" :0,
		"last_star" :0,
		"cost_star" :0,
		"energy_buf" :0,
		"gift_bought":[]int{},
		"collect":[]int{},
		"map_gift":[]int{},
		"medal_lv" :0,
		"icon" :3,
		"stars" :0,
		"nick": "冒险者",
		"level" :0,
		"exp" :0,
		"comCardIndex" :0,
		"bind" :0,
		"exchange" :0,
		"skin":"",
	}

	return retMap
}


func addUserData(ret map[string] interface{})  {
	ret["items"]=[]map[string] int{
		map[string] int{"uid":2786,"id":910,"num":3}}
	ret["cur_level"]=1
	ret["reward_id"]=0
	ret["gift"]=[]int{}
	ret["gift_back_time"]=0
	ret["pick_reward_id"]=0
	ret["reward_double"]=0
	ret["pig_card"]=0
}

// {"ret":0,"user":{"uid":2786,"coin":0,"energy":15,
// "max_energy":15,"last_recover":1491965576,"day_recover":0,
// "last_star":0,"cost_star":0,"energy_buf":0,"gift_bought":[],
// "collect":[],"map_gift":[],"medal_lv":0,
// "icon":3,"stars":0,"nick":"冒险者2786","level":0,
// "exp":0,"comCardIndex":0,"bind":0,"exchange":0,"skin":""},
// "blocks":[],"items":[],"cur_level":0}