function isArray(object){
	return object && typeof object ==='object' && Array == object.constructor
}

function MergeMoreServer(mergeServerArr,mergeServerStrArr,resultString,preString) {
	if (!isArray(mergeServerArr)){
		print("args mergeServerArr type is wrong!!")//检查args
		return 0
	}
  
	if (!isArray(mergeServerStrArr)){
		print("args mergeServerStrArr type is wrong!!")//检查args
		return 0
	}

	for (var i=0;i<mergeServerArr.length;i++){
		mergeServerArr[i]=NumberInt(mergeServerArr[i])
	}

	if(typeof(resultString) != "string"){
		print("args resultString type is wrong!!")//检查args
		return 0
	}

	if(resultString.length==0){
		print("args resultString content is empty!!")//检查args
		return 0
	}

	if(resultString === mergeServerStrArr[0]){
		print("args resultString content is wrong!!")//检查args
		return 0
	}

	if(typeof(preString) != "string"){
		print("args preString type is wrong!!")//检查args
		return 0
	}

	if(preString != "" && resultString.indexOf(preString) == 0){
		print("args resultString format is wrong!!")//检查args
		return 0
	}

	if(mergeServerArr.length<2){
		print("mergeServerArr length is wrong!!")//检查args
		return 0
	}

	if(mergeServerArr.length!=mergeServerStrArr.length){
		print("mergeServerArr or mergeServerStrArr length is wrong!!")//检查args
		return 0
	}

	var conn = new Mongo("mongodb://192.168.1.35:27017/admin");

	var dbAccount = conn.getDB(preString+"account");
	var checkServers=dbAccount.getCollection('server').find({"_id" : {$in:mergeServerArr}}).toArray()
	//printjson(checkServers);

	if(mergeServerArr.length != checkServers.length){
		print("mergeServerArr content is wrong!!because checkServers num is wrong!!!")//检查args
		return 0
	}

	var nowtime=new Date()
	var nowtimestamp = Date.parse(nowtime)/1000;

	for(var checkNodeIndex in checkServers){
		var checkNode=checkServers[checkNodeIndex]
		//printjson(checkNode);

		if(checkNode["time"]!=null){
			var tmpT=Date.parse(checkNode["time"])/1000
			//print(tmpT);
			if(tmpT>nowtimestamp){
				print("serverid "+checkNode["_id"]+" this server no open")
				return 0//此服没有开过
			}
		}

		if(checkNode["mergeid"]!=null && checkNode["mergeid"]>0){
			print("serverid "+checkNode["_id"]+" this server old!!")
			return 0//此服已经被合并过
		}
	}


	var nGameTo=mergeServerArr[0]
	var strGameTo=preString+resultString
	var dbGameTo = conn.getDB(strGameTo);

	var nMailTimeout=1296000


	for (var i=0;i<mergeServerArr.length;i++){
		var nGameFrom=mergeServerArr[i]
		var strGameFromTmp=""
		if (preString != "" && mergeServerStrArr[i].indexOf(preString) == 0) {
			strGameFromTmp=mergeServerStrArr[i]
		}else{
			strGameFromTmp=preString+mergeServerStrArr[i]
		}
		var strGameFrom=strGameFromTmp
		var dbGameFrom = conn.getDB(strGameFrom);


		var dbGameFromUsers=dbGameFrom.getCollection('users').find({})
		while(dbGameFromUsers.hasNext()) {
			var userInfo = dbGameFromUsers.next();

			if(userInfo["roles"].length<1) {
				continue
			}

			if (userInfo.logouttime>0 && nowtimestamp-userInfo.logouttime>1296000 && userInfo["roles"][0]["level"]<50) {
				continue
			}

			userInfo["mythicaltime"]=NumberLong(1);userInfo["onlinerectime"]=NumberLong(1)
			userInfo["togetherbuytime"]=NumberLong(1)

			dbGameTo.getCollection('users').update({_id:NumberInt(userInfo["_id"])},userInfo,true,false)
		}

		var dbGameFromGangs=dbGameFrom.getCollection('gangs').find({})
		while(dbGameFromGangs.hasNext()) {
			var gangInfo = dbGameFromGangs.next();

			dbGameTo.getCollection('gangs').update({_id:NumberInt(gangInfo["_id"])},gangInfo,true,false)
		}

		var dbGameFromMails=dbGameFrom.getCollection('mail').find({"state" : {$lt:1}})
		while(dbGameFromMails.hasNext()) {
			var mailInfo = dbGameFromMails.next();

			if (mailInfo["state"] > 0) {
				continue
			}

			if (mailInfo["time"]+nMailTimeout < nowtimestamp) {
				continue
			}

			dbGameTo.getCollection('mail').update({_id:mailInfo["_id"]},mailInfo,true,false)
		}


		var dbGameFromCounters=dbGameFrom.getCollection('counters').find({})
		while(dbGameFromCounters.hasNext()) {
			var counterInfo = dbGameFromCounters.next();

			if (counterInfo._id=="users"){
				dbGameTo.getCollection('counters').update({_id:"users"}, {$inc:{"seq" : NumberInt(counterInfo["seq"])}},true,false)
			}

			if (counterInfo._id=="gangs"){
				dbGameTo.getCollection('counters').update({_id:"gangs"}, {$inc:{"seq" : NumberInt(counterInfo["seq"])}},true,false)
			}
		}

		if (nGameFrom != nGameTo){
			dbAccount.getCollection('server').update({_id:NumberInt(nGameFrom)}, {$set:{"mergeid" : NumberInt(nGameTo)}},false,false)
		}else{
			dbAccount.getCollection('server').update({_id:NumberInt(nGameFrom)}, {$set:{"rundbname" : strGameTo}},false,false)			
		}

		dbAccount.getCollection('server').update({"mergeid" : NumberInt(nGameFrom)}, {$set:{"mergeid" : NumberInt(nGameTo)}},false,true)
	}

	dbAccount.getCollection('mergeServerlog').update({_id:NumberInt(nowtimestamp)}, {$set:{"mergeServerArr" : mergeServerArr,"mergeServerStrArr" : mergeServerStrArr,"resultString":resultString,"preString":preString,"datetime":nowtime}},true,false)

	return 1
}

//MergeServer(mergeServerArr,mergeServerStrArr,resultString,preString)
//mergeServerArr为[toGameid,fromGameids...]
//mergeServerStrArr为[toGamestr,fromGamestrs...]
//resultString为合并后的目标id服新的库名
//preString为渠道前缀如""，"yyb-"，"iwy-"
/*
var ret=MergeMoreServer([6,66,666],["game6","game66","game666"],"h-game6","sym-")
if (ret==1){
	print("ok")
}else{
	print("no ok")
}
*/


function CheckServerArr(mergeServerArr,mergeServerStrArr,mergeServerArrOut,mergeServerStrArrOut,preString) {
	var conn = new Mongo("mongodb://127.0.0.1:27017/admin");

	var dbAccount = conn.getDB(preString+"account");
	var checkServers=dbAccount.getCollection('server').find({"_id" : {$in:mergeServerArr}}).sort({"_id" :1}).toArray()
	//printjson(checkServers);

	for (var checkNodeIndex=0; checkNodeIndex<checkServers.length; checkNodeIndex++) {
		var checkNode=checkServers[checkNodeIndex]
		//printjson(checkNode)

		if(checkNode["mergeid"] == null){
			mergeServerArrOut.push(mergeServerArr[checkNodeIndex])

			if(checkNode["rundbname"] != null){
				mergeServerStrArrOut.push(checkNode["rundbname"])
			}else{
				mergeServerStrArrOut.push(mergeServerStrArr[checkNodeIndex])
			}
		}else{
			if(checkNode["mergeid"] == 0){
				mergeServerArrOut.push(mergeServerArr[checkNodeIndex])

				if(checkNode["rundbname"] != null){
					mergeServerStrArrOut.push(checkNode["rundbname"])
				}else{
					mergeServerStrArrOut.push(mergeServerStrArr[checkNodeIndex])
				}
			}
		}
	}
}

function ExecMS(startI,endI,numI,preString){
	if (startI <= 0){
		print("startI <= 0!!")
		return
	}

	if (endI-startI <= 0){
		print("endI <= startI!!")
		return
	}

	if (numI <= 1){
		print("numI <= 1!!")
		return
	}

	print("[开始服务区id 结束服务区id]:["+startI+", "+endI+"]",",每"+numI+"个合成一个新服")

	var doI=Math.ceil((endI-(startI-1))/numI)

	for (j=0;j<doI;j++){
		var arrayObj1 = new Array();
		var arrayObj2 = new Array();
		var arrayObj3 = new Array();
		var arrayObj4 = new Array();

		for (i=0;i<numI;i++){
			var tmp = startI+j*numI+i
			if (tmp>=startI && tmp<=endI){
				arrayObj1.push(tmp)
				arrayObj2.push("game"+tmp)
			}
		}
/*
		printjson(arrayObj1)
		printjson(arrayObj2)
		print("h-"+arrayObj2[0])
*/
		CheckServerArr(arrayObj1,arrayObj2,arrayObj3,arrayObj4,preString)

		if(arrayObj3.length!=arrayObj4.length){
			continue
		}

		if(arrayObj3.length==0){
			continue
		}

		printjson(arrayObj3)
		printjson(arrayObj4)
		print("h-"+arrayObj4[0])


		var ret=0
		ret=MergeMoreServer(arrayObj3,arrayObj4,"h-"+arrayObj4[0],preString)

		if (ret==1){
			print("ok")
		}else{
			print("no ok")
		}
	}
}

//ExecMS(331,340,10,"")





