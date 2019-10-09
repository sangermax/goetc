#ifndef BLSDK_H
#define BLSDK_H

#include "pubhead.h"

#  ifdef WIN32
#    define DECL_EXPORT     __declFTCc(dllexport)
#    define DECL_IMPORT     __declFTCc(dllimport)
#  elif defined(linux)
#    define DECL_EXPORT     __attribute__((visibility("default")))
#    define DECL_IMPORT     __attribute__((visibility("default")))
#  endif

#if defined(BLSDK_LIBRARY)
#  define BLSDKSHARED_EXPORT DECL_EXPORT
#else
#  define BLSDKSHARED_IMPORT DECL_IMPORT
#endif

#ifdef __cplusplus
extern "C" {
#endif

#ifdef BLSDK_LIBRARY
#define API BLSDKSHARED_EXPORT
#else
#define API BLSDKSHARED_IMPORT
#endif

    /*回调函数声明*/
    typedef void(*pReceivedMsgCallback_t)(const char * jsonMsg);

    /**
    *@brief 注册回调函数
    *@param 用户回调函数
    *@return 0 注册成功,-1 注册失败
    */
    API int RegisterCallbackFunction(pReceivedMsgCallback_t callback);

    /**
    *@brief 接收普通消息,从传输介质中接收到的数据都应该调用此函数处理.
    *@param dataBufIn 接收到的数据缓冲区
    *@param dataLen 接收到的数据长度
    *@return 无
    */
    API void ReceiveMessages(unsigned char* dataBufIn,int dataLenIn);

    /**
    *@brief 接收基于串口通信协议的消息,从传输介质中接收到的数据都应该调用此函数处理.
    *@param dataBufIn 接收到的数据缓冲区
    *@param dataLen 接收到的数据长度
    *@return 无
    */
    API void ReeiveSerialMessages(unsigned char* dataBufIn, int dataLenIn);

    /**
    *@brief 生成基于串口协议打包的数据.
    *@param dataBufIn 生成的普通消息包
    *@param dataLenIn 普通消息包的长度
    *@param dataBufOut 打包好的串口协议包.
    *@param dataLenOut 串口协议数据包的长度.
    *@return 无
    */
    API void GenerateMsgBySerialProtocol(unsigned char *dataBufIn,int dataLenIn,unsigned char *dataBufOut,int &dataLenOut);


    /**
    *@brief 获取心跳响应数据包
    *@param buf[OUT]输出指向内部静态数据缓冲区的指针,里面为实际的LLRP消息数据.
    *@param buflen[OUT]输出数据长度.
    *@return 无
    */
    API void GenerateKeepaliveAckData(unsigned char** buf, int& buflen);


    /**
    *@brief 获取断链请求消息数据包
    *@param buf[OUT]输出指向内部静态数据缓冲区的指针,里面为实际的LLRP消息数据.
    *@param buflen[OUT]输出数据长度.
    *@return 无
    */
    API void GenerateDisconnectData(unsigned char** buf,int& buflen);

    /**
    *@brief 获取断链应答消息数据包
    *@param buf[OUT]输出指向内部静态数据缓冲区的指针,里面为实际的LLRP消息数据.
    *@param buflen[OUT]输出数据长度.
    *@return 无
    */
    API void GenerateDisconnectAckData(unsigned char** buf, int& buflen);


    /**
    *@brief 获取能力获取消息数据包
    *@param buf[OUT]输出指向内部静态数据缓冲区的指针,里面为实际的LLRP消息数据.
    *@param buflen[OUT]输出数据长度.
    *@param type 请求的能力类型:0(ALL),1:(GenaralCapabilities),2:(CommunicationCapabilities),
            3:(FTCcCapabilities),4:(RfCapabilities),5:(AirProtocolCapabilities).
    *@return 无
    */
    API void GenerateGetDeviceCapabilitiesData(unsigned char** buf,int &buflen,unsigned char type);

#pragma region 标签选择
    /*********************************标签选择**********************************/

    /**
    *@brief 生成开始读卡消息.
    *@param buf[OUT]输出指向内部静态数据缓冲区的指针,里面为实际的LLRP消息数据.
    *@param buflen[OUT]输出数据长度.
    *@param antIds 天线ID数组.数组大小为4.选择相应的天线将相应的位置1.例如选择天线1.3.则antIds[0]=1.antIds[2]=1.
    *@param channelId 频点ID.(取值范围0-19)
    *@param rfPower 射频功率DB.(取值范围1-15)
    *@param startTriggerType //开始触发类型。0：无触发条件，仅当下发StartSelectFTCc 消息时触发；1：立即触发；2：周期触发；3：GPI 触发。(此参数填1)
    *@param selectType 标签操作类型:0-盘点.1-读.
    *@param readBankType 读分区类型(0-下半区,1-上半区,2-全区).(此参数填2)
    *@param membankIds 内存区域ID数组.数组大小为6.每个元素分别代表USER0-USER5.选择相应的内存区域参数将响应的位置1.例如选择USER0.USER5.则membankIds[0]=1,membankIds[4] = 1.
    *@param isNeedPersistance 是否需要持久化.0-不需要.1-需要.
    *@param isNeedTimmerRead 是否需要定时读卡.0-不需要.1-需要.
    *@param timmerReadMillonSec 定时读卡的时间(毫秒),如果需要定时读卡.则需设置该参数.
    */
    API void GenerateStartRead
    (
        unsigned char** buf, int &buflen,unsigned char * antIds,
        unsigned char channelId, unsigned char rfPower, unsigned char startTriggerType, unsigned char selectType, unsigned char readBankType, unsigned char * membankIds,
        unsigned char isNeedPersistance, unsigned char isNeedTimmerRead, unsigned int timmerReadMillonSec
    );
    /**
    *@brief 生成停止读卡消息.
    *@param buf[OUT]输出指向内部静态数据缓冲区的指针,里面为实际的LLRP消息数据.
    *@param buflen[OUT]输出数据长度.
    */
    API void GenerateStopRead(unsigned char** buf, int &buflen);

#pragma endregion



#pragma region 标签操作

    /*********************************标签操作**********************************/
    /**
    *@brief 生成开始写卡消息.
    *@param buf[OUT]输出指向内部静态数据缓冲区的指针,里面为实际的LLRP消息数据.
    *@param buflen[OUT]输出数据长度.
    *@param membankId - 内存区域ID.(USER0-USER5)
    *@param antId - 天线ID.(1-4)
    *@param data - 写入的数据数组(注:该数组元素为字的长度,最大可写入9个字的长度.)
    *@param dataLen - 写入的数据长度(max 9 word).
    */
    API void GenerateStartWrite(unsigned char** buf, int &len,unsigned char membankId, unsigned char antId, unsigned short *data, int dataLen);



    /**
    *@brief 生成写卡匹配读消息.在下发写卡指令成功后,需要下发该消息匹配读,将写入的数据读出.
    *@param buf[OUT]输出指向内部静态数据缓冲区的指针,里面为实际的LLRP消息数据.
    *@param buflen[OUT]输出数据长度.
    *@param antId - 天线ID.
    */
    API void GenerateStartWriteThenRead(unsigned char** buf, int &len, unsigned char antId);


    /**
    *@brief 生成停止写卡消息.调用此函数前,应调用GenerateStopRead删除所有读规则.
    *@param buf[OUT]输出指向内部静态数据缓冲区的指针,里面为实际的LLRP消息数据.
    *@param buflen[OUT]输出数据长度.
    */
    API void GenerateStopWrite(unsigned char** buf, int &len);

    /**
    *@brief 生成断链缓存应答消息.收到断链缓存消息后,应当回复应答消息.
    *@param buf[OUT]输出指向内部静态数据缓冲区的指针,里面为实际的LLRP消息数据.
    *@param buflen[OUT]输出数据长度.
    *@param sequenceID 接收到的缓存序号
    */
    API void GenerateCachedSelectAccessReportAck(unsigned char** buf, int &len,int sequenceID);

#pragma endregion


#pragma region 设备配置
    /**
    *@brief BellonConfigAddCommunicationLinkToBuf 添加TCP客户端(设备是Client)通信链路
    *@param buf[OUT]输出指向内部静态数据缓冲区的指针,里面为实际的LLRP消息数据.
    *@param buflen[OUT]输出数据长度.
    *@param IpAddress-IP地址字符串:例如:"192.168.1.1".(这里需要开发者自己检查IP地址是否正确)
    *@param port-端口(1024-65535)
    *@param hasKeepalive 心跳类型 0-无心跳,1-周期心跳.
    *@param periodicValue 心跳间隔时间:ms
    */
    API void GenerateSetTCPClientCommunicationLink(unsigned char ** buf, int & bufLen, char * ip, int port, unsigned char hasKeepalive, int periodicValue);


    /**
    *@brief BellonConfigAddCommunicationLinkToBuf 添加TCP服务端(设备是Server)通信链路
    *@param buf[OUT]输出指向内部静态数据缓冲区的指针,里面为实际的LLRP消息数据.
    *@param buflen[OUT]输出数据长度.
    *@param port-监听端口(1024-65535)
    *@param hasKeepalive 心跳类型 0-无心跳,1-周期心跳.
    *@param periodicValue 心跳间隔时间:ms
    */
    API void GenerateSetTCPServerCommunicationLink(unsigned char ** buf, int & bufLen,int port, unsigned char hasKeepalive, int periodicValue);


    /**
    *@brief GenerateConfigDeviceIP 配置设备ip地址(只支持ipv4)
    *@param buf-输入消息数据缓冲区,该函数会填充该缓冲区.建议大小:1024以上
    *@param bufLen-输入缓冲区内数据的长度,该函数会填充该数据.
    *@param ipAddr-ip地址字符串,例如:"192.168.0.20".
    *@param mask - 子网掩码字符串,例如:"255.255.255.0".
    *@param gateWayAddr - 网关地址字符串,例如"192.168.1.1".
    *@param dnsAddr - DNS地址字符串,例如"8.8.8.8".
    *@param isdhcp - 是否启用DHCP,0-不启用,1-启用.建议参数:0.
    *@return
    */
    API void GenerateConfigDeviceIP(unsigned char ** buf, int & bufLen, const char* ipAddr, const char* mask,const char* gateWayAddr, const char* dnsAddr,
        unsigned char isdhcp);

    /**
    *@brief GenerateConfigNTP 配置设备NTP时间同步服务器
    *@param buf[OUT]输出指向内部静态数据缓冲区的指针,里面为实际的LLRP消息数据.
    *@param buflen[OUT]输出数据长度.
    *@param ipAddr-时间同步服务器IP地址,例如"184.324.12.32"
    *@param ntpPeriodHour 同步周期.单位(小时)
    */
    API void GenerateConfigNTP(unsigned char** buf, int& bufLen, const char* ipAddr, unsigned int ntpPeriodHour);


    /**
    *@brief GenerateConfigDeviceLocation 配置设备地理位置
    *@param buf[OUT]输出指向内部静态数据缓冲区的指针,里面为实际的LLRP消息数据.
    *@param buflen[OUT]输出数据长度.
    *@param longitude - 经度值,乘以1000000表示,正数表示东经,负数表示西经.
    *@param latitude -  纬度值,乘以1000000表示,正数表示北纬,负数表示南纬.
    *@param altitude -  海拔值,乘以100表示.
    */
    API void GenerateConfigDeviceLocation(unsigned char** buf, int& bufLen, signed int longitude, signed int latitude, signed int altitude);

    /**
    *@brief GenerateRebootDevice 重启设备
    *@param buf[OUT]输出指向内部静态数据缓冲区的指针,里面为实际的LLRP消息数据.
    *@param buflen[OUT]输出数据长度.
    */
    API void GenerateRebootDevice(unsigned char** buf, int&bufLen);

    /**
    *@brief GenerateGetDeviceConfig 获取设备配置
    *@param buf[OUT]输出指向内部静态数据缓冲区的指针,里面为实际的LLRP消息数据.
    *@param buflen[OUT]输出数据长度.
    *@param requestID-请求配置id,取值范围(0-11)0:所有配置,1:设备唯一标识,2:设备事件通知标志,3:通信配置,4:告警配置,5:天线属性配置,
    *				6:天线协议配置,7:调制深度配置,8,标签选择报告规则配置,9:标签操作报告规则配置,10:设备位置配置,11:安全模块配置.
    *@param antId-天线ID,取值范围(0-4),0表示所有天线.1-4表示天线ID编号
    */
    API void GenerateGetDeviceConfig(unsigned char** buf, int& bufLen, unsigned char requestID, unsigned char antID);


    /**
    *@brief GenerateGetVersionInfo 获取设备版本信息
    *@param buf[OUT]输出指向内部静态数据缓冲区的指针,里面为实际的LLRP消息数据.
    *@param buflen[OUT]输出数据长度.
    *@param versionType-获取版本的类型,取值范围(0-4)0:读写器(设备)BOOT版本,1:读写器系统版本,2:安全模块系统版本(由安全模块硬件版本号,
    安全模块控制芯片boot版本号以及安全模块控制芯片用户程序版本号组成);3:安全芯片系统版本(由加密模块boot版本号,加密模块1用户程序版本
    号以及加密模块2用户程序版本号组成);4:安全模块密钥版本.
    */
    API void GenerateGetVersionInfo(unsigned char** buf, int& bufLen, unsigned char versionType);


    /**
    *@brief GenerateCurrentTime 设备校准时间:发送当前PC的时间戳到设备.
    *@param buf[OUT]输出指向内部静态数据缓冲区的指针,里面为实际的LLRP消息数据.
    *@param buflen[OUT]输出数据长度.
    */

    API void GenerateCurrentTime(unsigned char** buf, int& buflen);

    /**
    *@brief GenerateGetDeviceTime 获取设备当前时间戳
    *@param buf[OUT]输出指向内部静态数据缓冲区的指针,里面为实际的LLRP消息数据.
    *@param buflen[OUT]输出数据长度.
    */
    API void GenerateGetDeviceTime(unsigned char** buf, int& buflen);


    /**
    *@brief BellonConfigAlarmToBuf 配置告警信息
    *@param buf[OUT]输出指向内部静态数据缓冲区的指针,里面为实际的LLRP消息数据.
    *@param buflen[OUT]输出数据长度.
    *@param alarmMask 告警级别掩码,从低到高每位表示对应的告警级别是否上报.0:不上报,1:上报.
    *			   其中第0位表示的的告警级别为致命,其他告警级别以此类推.
    *@tempreatureMax  最大温度告警阀值
    *@tempreatureMin  最低温度告警阀值
    */
    API void GenerateConfigAlarm(unsigned char** buf, int& buflen, unsigned char alarmMask, signed char tempreatureMax, signed char tempreatureMin);

    /**
    *@brief GenerateDeleteAlarm 删除告警信息
    *@param buf[OUT]输出指向内部静态数据缓冲区的指针,里面为实际的LLRP消息数据.
    *@param buflen[OUT]输出数据长度.
    *@param alarmId  删除的告警编号
    */
    API void GenerateDeleteAlarm(unsigned char** buf, int& buflen, unsigned int alarmId);


    /**
    *@brief GenerateAlarmSync 同步告警信息
    *@param buf[OUT]输出指向内部静态数据缓冲区的指针,里面为实际的LLRP消息数据.
    *@param buflen[OUT]输出数据长度.
    *@param syncSeq -同步序列号(自增变量)
    *@param alarmIds  同步的告警编号数组
    *@param alarmIdCount -同步告警编号数组长度
    */
    API void GenerateAlarmSync(unsigned char** buf, int& buflen, unsigned int syncSeq, unsigned int* alarmIds, int alarmIdCount);

#pragma endregion

#pragma region 设备日志管理
    /**
    *@brief GenerateUploadDeviceLog 请求上报设备日志消息
    *@param buf[OUT]输出指向内部静态数据缓冲区的指针,里面为实际的LLRP消息数据.
    *@param buflen[OUT]输出数据长度.
    *@param uploadType -上报类型(0:按日志条数上传,上传最新的日志;1:按日志时间上传;2:定时上传)
    *@param uploadCount - uploadType = 0 时,上传的日志条数.
    *@param ms  -uploadType = 2 时,定时上传的毫秒数.
    *@param startTimestampMs - uploadType = 1 时,开始时间戳.(单位:ms 注意.需传入UTC标准时间戳)
    *@param stopTimestampMs - uploadType = 1 时,停止时间戳.(单位:ms 注意.需传入UTC标准时间戳)
    */
    API void GenerateUploadDeviceLog(unsigned char** buf, int& buflen, unsigned char uploadType, unsigned int uploadCount, unsigned int ms, unsigned long long startTimestampMs, unsigned long long stopTimestampMs);



    /**
    *@brief GenerateUploadDeviceLogConfirm 设备日志上传应答确认
    *@param buf[OUT]输出指向内部静态数据缓冲区的指针,里面为实际的LLRP消息数据.
    *@param bufLen[OUT]输出数据长度.
    *@param sequenceId -日志序号
    */
    API void GenerateUploadDeviceLogConfirm(unsigned char** buf, int&bufLen, unsigned short sequenceId);


    /**
    *@brief GenerateClearDeviceLog 设备日志清空请求消息
    *@param buf[OUT]输出指向内部静态数据缓冲区的指针,里面为实际的LLRP消息数据.
    *@param bufLen[OUT]输出数据长度.
    */
    API void GenerateClearDeviceLog(unsigned char** buf, int&bufLen);


    /**
    *@brief GenerateGetDeviceLogCount 获取设备日志数量请求消息
    */
    API void GenerateGetDeviceLogCount(unsigned char** buf, int&bufLen);



#pragma endregion



#ifdef __cplusplus
}
#endif

#endif
