package main
import (
        "encoding/json"
        "github.com/hyperledger/fabric/common/util"
        "fmt"
        "strings"
        "strconv" 
        "github.com/hyperledger/fabric/core/chaincode/shim"
        pb "github.com/hyperledger/fabric/protos/peer"
)
const success="success"
// channelId:="mychannel"
// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}
//持仓表
type SSEHold struct{
    ObjectType string `json:"docType"`
    AccId string `json:"accId"`
    ProductCode string `json:"productCode"`
    HoldNum int `json:"holdNum"`
    FrozenSecNum int `json:"frozenSecNum"`
}
type AcctAsset struct{
    ObjectType string `json:"docType"`
    AcctId string `json:"acctId"`
    AvaMoney int `json:"avaMoney"`
}
type ContractHold struct{
    ObjectType string `json:"docType"`
    ContractN string `json:"contractN"`
    ContractCode string `json:"contractCode"`
    ContractStatus string `json:"contractStatus"`//"0"代表等待开始，"1"代表开始合约，"2"代表已经开始,
//"3"代表行权，"4"代表已终结,"5"代表互操作,"6"代表待联通,"7"代表联通,"8"代表解质押
    ContractCCID string `json:"contractCCID"`//合约名
    ContractFunctionName string `json:"contractFunctionName"`
    AccIdA string `json:"accIdA"`//合约关联方A
    AccIdB string `json:"accIdB"`
    AccIdC string `json:"accIdC"`
    AccIdD string `json:"accIdD"`
    AccIdE string `json:"accIdE"`
    CorContractNA string `json:"corContractNA"`//关联合约
    CorContractNB string `json:"corContractNB"`
    CorContractNC string `json:"corContractNC"`
    TransType string `json:"transType"`//"0"代表不可转让，"1"代表可转让
    LastSwapTime string `json:"lastSwapTime"`
}
//账户流水表
type AccFlow struct{
    ObjectType string `json:"docType"`
    AccFlowId string `json:"accFlowId"`//账户流水编号
    AccId string `json:"accId"`
    AssetId string `json:"assetId"`
    AssetNum string `json:"assetNum"`
    SType string `json:"sType"`//"0"代表增加,"1"代表减少,"2"代表锁定,"3"代表解锁
    ContractN string `json:"contractN"`
    Time string `json:"time"`
}
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response  {
        fmt.Println("########### example_cc Init ###########")
    
    return shim.Success(nil)
}
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface) pb.Response {
        return shim.Error("Unknown supported call")
}
// Transaction makes payment of X units from A to B
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
        fmt.Println("########### example_cc Invoke ###########")
    function, args := stub.GetFunctionAndParameters()
    
    if function != "invoke" {
                return shim.Error("Unknown function call")
    }

    if len(args) < 2 {
        return shim.Error("Incorrect number of arguments. Expecting at least 2")
    }
    //设计合约
    if args[0]=="blocktradetwo"{
        return t.blocktradetwo(stub,args)
    }
    return shim.Error("Unknown action, check the first argument, must be one of 'delete', 'query', or 'move'")
}
//设计合约
//bigDealTwo,contractN,actionType,accId,pw,accIdN,corContractNAHoldStr,
//(accFlowId,acctAssetA,acctAssetB,sseHoldA,sseHoldB),channelId,chaincodeToCall
func (t *SimpleChaincode) blocktradetwo(stub shim.ChaincodeStubInterface,args []string)pb.Response{
    if len(args)!=10{
        return shim.Error("args numbers is wrong")
    }
    if args[1]==""||args[2]==""||args[3]==""||args[4]==""||args[5]==""{
        return shim.Error("args can not be nil")
    }
    contractNStr:=args[1]
    actionType:=args[2]
    accIdStr:=args[3]
    pwStr:=args[4]
    accIdNStr:=args[5]
    corContractNAHoldStr:=args[6]
    allData:=args[7]
    channelId:=args[8]
    chaincodeToCall:=args[9]
    //账户流水
    var accFlowAAsset,accFlowBAsset,accFlowAHold,accFlowBHold AccFlow   
    //资金账户
    var acctAssetA,acctAssetB AcctAsset
    //持仓账户
    var sseHoldA,sseHoldB SSEHold
    //获得合约持仓
    contractHold,result:=getContractHold(stub,contractNStr,chaincodeToCall,channelId)
    if result!=success{
        return shim.Error(result)
    }
    var accFlowIdStr string
    isTermination:=true
    if corContractNAHoldStr!=""{
        fmt.Println("bigdealtwo2 111 contractHold.CorContractNB is not nil")
        //获得关联合约持仓
        var corContractNAHold ContractHold
        err:=json.Unmarshal([]byte(corContractNAHoldStr),&corContractNAHold)
        if err!=nil{
            return shim.Error("corContractNAHold Unmarshal failed")
        }
        //(accFlowId,acctAssetA,sseHoldA)构成allData
        //获得最新的账户流水编号
        dataArr:=strings.Split(allData,";")
        accFlowIdStr=dataArr[0]
        accFlowId,err:=strconv.Atoi(accFlowIdStr)
        if err!=nil{
            return shim.Error("accFlowId Atoi failed")
        }
        //获得B账户的资金账户和持仓账户
        err=json.Unmarshal([]byte(dataArr[1]),&acctAssetB)
        if err!=nil{
            return shim.Error("acctAssetB Unmarshal failed")
        }
        //有可能只传进来资金账户而没有持仓账户
        if len(dataArr)!=3{
            sseHoldB,result=getSSEHoldByAccAndProduct(stub,contractHold.AccIdB,"sh0003",chaincodeToCall,channelId)
            if result!=success{
                return shim.Error(result)
            }
        }else{
            err=json.Unmarshal([]byte(dataArr[2]),&sseHoldB)
            if err!=nil{
                return shim.Error("sseHoldB Unmarshal failed")
            }
        }
        //获得A账户的资金账户和持仓账户
        //获得账户A的资金账户
        acctAssetA,result=getAcctAsset(stub,contractHold.AccIdA,chaincodeToCall,channelId)
        if result!=success{
            return shim.Error(result)
        }
        //获得账户A的持仓账户
        sseHoldA,result=getSSEHoldByAccAndProduct(stub,contractHold.AccIdA,"sh0003",chaincodeToCall,channelId)
        if result!=success{
            return shim.Error(result)
        }
        if actionType=="7"&&corContractNAHold.ContractStatus=="6"{   
            fmt.Println("bigdealtwo2 222 contractHold.CorContractNB is not nil")
            if acctAssetA.AvaMoney<2000{
                return shim.Error("acctIdA do not have enough money")
            }   
            //判断是否有足够持仓
            if sseHoldB.HoldNum<100{
                return shim.Error("acctIdB do not have enough stock")
            } 
            //A给B转钱
            acctAssetA.AvaMoney-=70000
            acctAssetB.AvaMoney+=70000
            //账户流水
            timeStr:=getCurrTime(stub,chaincodeToCall,channelId)
            accFlowAAsset=AccFlow{
                "accFlow",
                accFlowIdStr,
                contractHold.AccIdA,
                "money",
                "70000",
                "1",
                contractNStr,
                timeStr,
            }
            accFlowId=accFlowId+1
            accFlowIdStr=strconv.Itoa(accFlowId)
            accFlowBAsset=AccFlow{
                "accFlow",
                accFlowIdStr,
                contractHold.AccIdB,
                "money",
                "70000",
                "0",
                contractNStr,
                timeStr,
            }
            //B给A转券
            sseHoldB.HoldNum-=1000
            sseHoldA.HoldNum+=1000
            //账户流水
            accFlowId=accFlowId+1
            accFlowIdStr=strconv.Itoa(accFlowId)
            accFlowAHold=AccFlow{
                "accFlow",
                accFlowIdStr,
                contractHold.AccIdA,
                "sh0003",
                "1000",
                "0",
                contractNStr,
                timeStr,
            }
            accFlowId=accFlowId+1
            accFlowIdStr=strconv.Itoa(accFlowId)
            accFlowBHold=AccFlow{
                "accFlow",
                accFlowIdStr,
                contractHold.AccIdB,
                "sh0003",
                "1000",
                "1",
                contractNStr,
                timeStr,
            }
            //判断B是否有足够资金
            if acctAssetB.AvaMoney<70000{
                return shim.Error("acctB do not have enough money")
            }
            //判断A是否有足够券
            if sseHoldA.HoldNum<1000{
                return shim.Error("acctA do not have enough stock")
            }
            //切记
            contractHold.ContractStatus="6"
            //获得corContractNB的合约持仓
            if contractHold.CorContractNB!=""{
                fmt.Println("bigdealtwo2 333 contractHold.CorContractNB is not nil")
                isTermination=false
                //最后更新流水号
                accFlowId=accFlowId+1
                accFlowIdStr=strconv.Itoa(accFlowId)
                corContractHoldB,result:=getContractHold(stub,contractHold.CorContractNB,chaincodeToCall,channelId)
                if result!=success{
                    return shim.Error(result)
                }
                //构造调用参数
                contractHoldBytes,err:=json.Marshal(contractHold)
                if err!=nil{
                    return shim.Error("contractHoldBytes Marshal failed")
                }
                acctAssetABytes,err:=json.Marshal(acctAssetA)
                if err!=nil{
                    return shim.Error("acctAssetABytes Marshal failed")
                }
                sseHoldABytes,err:=json.Marshal(sseHoldA)
                if err!=nil{
                    return shim.Error("sseHoldABytes Marshal failed")
                }
                allData=accFlowIdStr+";"+string(acctAssetABytes)+";"+string(sseHoldABytes)
                //调用corContractHoldB中指定的函数,"7"代表联通
                invokeArgs:=util.ToChaincodeArgs("invoke",corContractHoldB.ContractFunctionName,contractHold.CorContractNB,"7",accIdStr,pwStr,"0",string(contractHoldBytes),allData,"mychannel",chaincodeToCall)
                response := stub.InvokeChaincode(corContractHoldB.ContractCCID, invokeArgs, channelId)
                if response.Status!=shim.OK{
                    errStr := fmt.Sprintf("getContractHold Failed 8888to invoke chaincode. Got error: %s", string(response.Payload))
                    fmt.Printf(errStr)
                    return shim.Error(errStr+chaincodeToCall+channelId+corContractHoldB.ContractCCID+corContractHoldB.ContractFunctionName+allData)
                }
            }
            contractHold.ContractStatus="4"
        }else{
            return shim.Error("unvalid Status")
        }
    }else{     
        fmt.Println("bigdealtwo2 444 contractHold.CorContractNB is not nil")
        var accIdN string
        if accIdNStr=="1"{
            accIdN=contractHold.AccIdA
        }else if accIdNStr=="2"{
            accIdN=contractHold.AccIdB
        }else if accIdNStr=="3"{
            accIdN=contractHold.AccIdC
        }else if accIdNStr=="4"{
            accIdN=contractHold.AccIdD
        }else if accIdNStr=="5"{
            accIdN=contractHold.AccIdE
        }
        if accIdStr!=accIdN{
            return shim.Error("accId is wrong"+accIdStr+","+accIdN)
        }
        //获得权限
        result=getPermission(stub,accIdStr,pwStr,chaincodeToCall,channelId)
        if result!=success{
            return shim.Error(result)
        }
        //获得最新的账户流水编号
        accFlowIdStr,result=getAccFlowId(stub,chaincodeToCall,channelId)
        if result!=success{
            return shim.Error(result)
        }
        accFlowId,err:=strconv.Atoi(accFlowIdStr)
        if err!=nil{
            return shim.Error("accFlowId Atoi failed")
        }
        //获得账户A的资金账户
        acctAssetA,result=getAcctAsset(stub,contractHold.AccIdA,chaincodeToCall,channelId)
        if result!=success{
            return shim.Error(result)
        }
        //获得B的资金账户
        acctAssetB,result=getAcctAsset(stub,contractHold.AccIdB,chaincodeToCall,channelId)
        if result!=success{
            return shim.Error(result)
        }
        //获得账户B的持仓账户
        sseHoldB,result=getSSEHoldByAccAndProduct(stub,contractHold.AccIdB,"sh0003",chaincodeToCall,channelId)
        if result!=success{
            return shim.Error(result)
        }
        //获得账户A的持仓账户
        sseHoldA,result=getSSEHoldByAccAndProduct(stub,contractHold.AccIdA,"sh0003",chaincodeToCall,channelId)
        if result!=success{
            return shim.Error(result)
        }
        if contractHold.ContractStatus=="0"&&actionType=="1"{
            fmt.Println("bigdealtwo2 555 contractHold.CorContractNB is not nil")
            if accIdStr!=contractHold.AccIdA{
                return shim.Error("accId is wrong")
            }       
            //判断资金是否够锁钱
            if acctAssetA.AvaMoney<70000{
                return shim.Error("accIdA avamoney is not enough")
            }
            //判断是否有足够的持仓数据
            if sseHoldB.HoldNum<1000{
                return shim.Error("accountB holdNum is not enough")
            }
            //A给B转钱
            acctAssetA.AvaMoney-=70000
            acctAssetB.AvaMoney+=70000
            //B给A转券
            sseHoldB.HoldNum-=1000
            sseHoldA.HoldNum+=1000
            //钱账户流水
            timeStr:=getCurrTime(stub,chaincodeToCall,channelId)
            accFlowAAsset=AccFlow{
                "accFlow",
                accFlowIdStr,
                contractHold.AccIdA,
                "money",
                "70000",
                "1",
                contractNStr,
                timeStr,
            }
            accFlowId=accFlowId+1
            accFlowIdStr=strconv.Itoa(accFlowId)
            accFlowBAsset=AccFlow{
                "accFlow",
                accFlowIdStr,
                contractHold.AccIdB,
                "money",
                "70000",
                "0",
                contractNStr,
                timeStr,
            }
            //券账户流水
            accFlowId=accFlowId+1
            accFlowIdStr=strconv.Itoa(accFlowId)
            accFlowAHold=AccFlow{
                "accFlow",
                accFlowIdStr,
                contractHold.AccIdA,
                "sh0003",
                "1000",
                "0",
                contractNStr,
                timeStr,
            }
            accFlowId=accFlowId+1
            accFlowIdStr=strconv.Itoa(accFlowId)
            accFlowBHold=AccFlow{
                "accFlow",
                accFlowIdStr,
                contractHold.AccIdB,
                "sh0003",
                "1000",
                "1",
                contractNStr,
                timeStr,
            }
            contractHold.ContractStatus="6"
            //调用关联合约corContractNB的持仓表里函数
            if contractHold.CorContractNB!=""{
                fmt.Println("bigdealtwo2 666 contractHold.CorContractNB is not nil")
                isTermination=false
                //最后更新账户流水号
                accFlowId=accFlowId+1
                accFlowIdStr=strconv.Itoa(accFlowId)
                corContractHoldB,result:=getContractHold(stub,contractHold.CorContractNB,chaincodeToCall,channelId)
                if result!=success{
                    return shim.Error(result)
                }
                //构造调用参数
                contractHoldBytes,err:=json.Marshal(contractHold)
                if err!=nil{
                    return shim.Error("contractHoldBytes Marshal failed")
                }
                acctAssetABytes,err:=json.Marshal(acctAssetA)
                if err!=nil{
                    return shim.Error("acctAssetABytes Marshal failed")
                }
                sseHoldABytes,err:=json.Marshal(sseHoldA)
                if err!=nil{
                    return shim.Error("sseHoldABytes Marshal failed")
                }

                allData=accFlowIdStr+";"+string(acctAssetABytes)+";"+string(sseHoldABytes)
                invokeArgs:=util.ToChaincodeArgs("invoke",corContractHoldB.ContractFunctionName,contractHold.CorContractNB,"7",accIdStr,pwStr,"0",string(contractHoldBytes),allData,"mychannel",chaincodeToCall)
                response := stub.InvokeChaincode(corContractHoldB.ContractCCID,invokeArgs,channelId)
                if response.Status!=shim.OK{
                    errStr := fmt.Sprintf("3 getContractHold Failed to invoke chaincode. Got error: %s", string(response.Payload))
                    fmt.Printf(errStr)
                    return shim.Error(errStr+chaincodeToCall+channelId+corContractHoldB.ContractCCID+";"+string(contractHoldBytes)+";"+corContractHoldB.ContractFunctionName+allData)
                }
            }
            contractHold.ContractStatus="4"
        }else{
            return shim.Error("unvalid Status")
        }       
    } 
    //保存合约最新的状态
    result=saveContractHold(stub,contractHold,contractNStr,chaincodeToCall,channelId)
    if result!=success{
            return shim.Error(result)
    }
    ///////////////////////////特别声明资金账户和持仓账户的更新应该在另一个chaincode里面?????
    //保存A账户资金账户和持仓账户
    if isTermination{
        fmt.Println("bigdealtwo2 777 contractHold.CorContractNB is not nil")
        result=saveAcctAsset(stub,acctAssetA,acctAssetA.AcctId,chaincodeToCall,channelId)
        if result!=success{
            return shim.Error(result)
        }
        result=saveSSEHoldByAccAndProduct(stub,sseHoldA,sseHoldA.AccId,"sh0003",chaincodeToCall,channelId)
        if result!=success{
            return shim.Error(result)
        }
        //保存最新的账户流水号
        result=saveAccFlowId(stub,accFlowIdStr,chaincodeToCall,channelId)
        if result!=success{
            return shim.Error(result)
        }
    }
    //保存B账户资金账户和持仓账户
    result=saveAcctAsset(stub,acctAssetB,acctAssetB.AcctId,chaincodeToCall,channelId)
    if result!=success{
        return shim.Error(result)
    }
    result=saveSSEHoldByAccAndProduct(stub,sseHoldB,sseHoldB.AccId,"sh0003",chaincodeToCall,channelId)
    if result!=success{
        return shim.Error(result)
    }
    //保存账户流水
    result=saveAccFlow(stub,accFlowAAsset,accFlowAAsset.AccFlowId,chaincodeToCall,channelId)
    if result!=success{
        return shim.Error(result)
    }
    result=saveAccFlow(stub,accFlowBAsset,accFlowBAsset.AccFlowId,chaincodeToCall,channelId)
    if result!=success{
        return shim.Error(result)
    }
    result=saveAccFlow(stub,accFlowAHold,accFlowAHold.AccFlowId,chaincodeToCall,channelId)
    if result!=success{
        return shim.Error(result)
    }
    result=saveAccFlow(stub,accFlowBHold,accFlowBHold.AccFlowId,chaincodeToCall,channelId)
    if result!=success{
        return shim.Error(result)
    }
    return shim.Success([]byte(success))
}
//获得系统时间
func getCurrTime(stub shim.ChaincodeStubInterface,chaincodeToCall,channelId string)string{
    invokeArgs:=util.ToChaincodeArgs("invoke","getCurrTime")
    response:= stub.InvokeChaincode(chaincodeToCall, invokeArgs, channelId)
    if response.Status!=shim.OK{
        errStr := fmt.Sprintf("getCurrTime Failed to invoke chaincode. Got error: %s", string(response.Payload))
        fmt.Printf(errStr)
        return errStr
    }
    return string(response.Payload)
}
//保存账户流水
func saveAccFlow(stub shim.ChaincodeStubInterface,accFlow AccFlow,accIdStr,chaincodeToCall,channelId string)string{
    accFlowBytes,err:=json.Marshal(accFlow)
    if err!=nil{
        return "accFlowBytes Marshal failed"
    }
    invokeArgs:=util.ToChaincodeArgs("invoke","saveAccFlowOut",accIdStr,string(accFlowBytes))
    response := stub.InvokeChaincode(chaincodeToCall, invokeArgs, channelId)
    if response.Status!=shim.OK{
        errStr := fmt.Sprintf("4getContractHold Failed to invoke chaincode. Got error: %s", string(response.Payload))
        fmt.Printf(errStr)
        return errStr
    }
    return success
}
//获得合约持仓
func getContractHold(stub shim.ChaincodeStubInterface,contractNStr,chaincodeToCall,channelId string)(ContractHold,string){
    var contractHold ContractHold
    invokeArgs:=util.ToChaincodeArgs("invoke","getContractHold",contractNStr)
    response := stub.InvokeChaincode(chaincodeToCall, invokeArgs, channelId)
    if response.Status!=shim.OK{
        errStr := fmt.Sprintf("5 getContractHold Failed to invoke chaincode. Got error: %s", string(response.Payload))
        fmt.Printf(errStr)
        return contractHold,errStr
    }
    contractHoldBytes:=response.Payload
    err:=json.Unmarshal(contractHoldBytes,&contractHold)
    if err!=nil{
        return contractHold,"contractHoldBytes Unmarshal failed"
    }
    return contractHold,success
}
//获得资金账户
func getAcctAsset(stub shim.ChaincodeStubInterface,accIdStr,chaincodeToCall,channelId string)(AcctAsset,string){
    var acctAsset AcctAsset
    invokeArgs:=util.ToChaincodeArgs("invoke","getAcctAsset",accIdStr)
    response:= stub.InvokeChaincode(chaincodeToCall, invokeArgs, channelId)
    if response.Status!=shim.OK{
        errStr := fmt.Sprintf("6 getContractHold Failed to invoke chaincode. Got error: %s", string(response.Payload))
        fmt.Printf(errStr)
        return acctAsset,errStr
    }
    acctAssetBytes:=response.Payload
    err:=json.Unmarshal(acctAssetBytes,&acctAsset)
    if err!=nil{
        return acctAsset,"acctAssetABytes Unmarshal failed"
    }
    return acctAsset,success
}
//保存资金账户
func saveAcctAsset(stub shim.ChaincodeStubInterface,acctAsset AcctAsset,accIdStr,chaincodeToCall,channelId string)string{
    acctAssetBytes,err:=json.Marshal(acctAsset)
    if err!=nil{
        return "acctAssetBytes Marshal failed"
    }
    invokeArgs:=util.ToChaincodeArgs("invoke","saveAcctAsset",accIdStr,string(acctAssetBytes))
    response:= stub.InvokeChaincode(chaincodeToCall, invokeArgs, channelId)
    if response.Status!=shim.OK{
        errStr := fmt.Sprintf("7getContractHold Failed to invoke chaincode. Got error: %s", string(response.Payload))
        fmt.Printf(errStr)
        return errStr
    }
    return success
}
//获得持仓账户
func getSSEHoldByAccAndProduct(stub shim.ChaincodeStubInterface,accIdStr,productCode,chaincodeToCall,channelId string)(SSEHold,string){
    var sseHold SSEHold 
    invokeArgs:=util.ToChaincodeArgs("invoke","getSSEHoldByAccAndProduct",accIdStr,productCode)
    response:= stub.InvokeChaincode(chaincodeToCall, invokeArgs, channelId)
    if response.Status!=shim.OK{
        errStr := fmt.Sprintf("A B lockMoney Failed to invoke chaincode. Got error: %s", string(response.Payload))
        fmt.Printf(errStr)
        return sseHold,errStr
    }
    sseHoldBytes:=response.Payload
    err:=json.Unmarshal(sseHoldBytes,&sseHold) 
    if err!=nil{
        return sseHold,"sseHoldBytes Unmarshal failed"
    }
    return sseHold,success
}
//保存持仓账户信息
func saveSSEHoldByAccAndProduct(stub shim.ChaincodeStubInterface,sseHold SSEHold,accIdStr,productCode,chaincodeToCall,channelId string)string{
    sseHoldBytes,err:=json.Marshal(sseHold)
    if err!=nil{
        return "sseHoldBytes marshal failed"
    }
    invokeArgs:=util.ToChaincodeArgs("invoke","saveSSEHoldByAccAndProduct",accIdStr,productCode,string(sseHoldBytes))
    response := stub.InvokeChaincode(chaincodeToCall, invokeArgs, channelId)
    if response.Status!=shim.OK{
        errStr := fmt.Sprintf("A B lockMoney Failed to invoke chaincode. Got error: %s", string(response.Payload))
        fmt.Printf(errStr)
        return errStr
    }
    return success
}
//保存合约最新的状态
func saveContractHold(stub shim.ChaincodeStubInterface,contractHold ContractHold,contractNStr,chaincodeToCall,channelId string)string{
    contractHoldBytes,err:=json.Marshal(contractHold)
    if err!=nil{
        return "contractHold marshal failed"
    }
    fmt.Println("bigdealtwo2"+";"+string(contractHoldBytes))
    invokeArgs:=util.ToChaincodeArgs("invoke","saveContractHold",contractNStr,string(contractHoldBytes))
    response := stub.InvokeChaincode(chaincodeToCall, invokeArgs, channelId)
    if response.Status!=shim.OK{
        errStr := fmt.Sprintf("7 Failed to invoke chaincode. Got error: %s", string(response.Payload))
        fmt.Printf(errStr)
        return errStr
    }
    return success
}
//保存最新的账户流水编号
func saveAccFlowId(stub shim.ChaincodeStubInterface,accFlowIdStr,chaincodeToCall,channelId string)string{
    invokeArgs:=util.ToChaincodeArgs("invoke","saveAccFlowId",accFlowIdStr)
    response:=stub.InvokeChaincode(chaincodeToCall, invokeArgs, channelId)
    if response.Status!=shim.OK{
        errStr := fmt.Sprintf("B saveAcctAsset Failed to invoke chaincode. Got error: %s", string(response.Payload))
        fmt.Printf(errStr)
        return errStr
    }
    return success
}
//获得最新的账户流水编号
func getAccFlowId(stub shim.ChaincodeStubInterface,chaincodeToCall,channelId string)(string,string){
    invokeArgs:=util.ToChaincodeArgs("invoke","getAccFlowId")
    response := stub.InvokeChaincode(chaincodeToCall,invokeArgs,channelId)
    if response.Status!=shim.OK{
        errStr := fmt.Sprintf("2 getContractHold Failed to invoke chaincode. Got error: %s", string(response.Payload))
        fmt.Printf(errStr)
        return "0",errStr
    }
    return string(response.Payload),success
}
//获得权限
func getPermission(stub shim.ChaincodeStubInterface,accIdStr,pwStr,chaincodeToCall,channelId string)string{
    invokeArgs:=util.ToChaincodeArgs("invoke","getAccPermissionOut",accIdStr,pwStr)
    response := stub.InvokeChaincode(chaincodeToCall, invokeArgs, channelId)
    if response.Status!=shim.OK{
        errStr := fmt.Sprintf(" 1 getContractHold Failed to invoke chaincode. Got error: %s", string(response.Payload))
        fmt.Printf(errStr)
        return errStr+"getPermission"
    }
    return success
}
func main() {
    err := shim.Start(new(SimpleChaincode))
    if err != nil {
        fmt.Printf("Error starting Simple chaincode: %s", err)
    }
}
