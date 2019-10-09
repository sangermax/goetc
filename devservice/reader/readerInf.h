#ifndef __READERINF_H__
#define __READERINF_H__

#ifdef __cplusplus
extern "C" {
#endif

int JT_OpenReader(char* sDevInfo,int iComID,int iBaudrate);
int JT_CloseReader();
//int JT_SetPCBaudrate(long iBaudrate);
//int JT_SetReaderBaudrate(long iBaudrate);
int JT_GetLastError(char*  sLastError);
//void JT_sErrorInfo(int ErrorCode, char*  sErrorInfo);
int JT_OpenCard();
int JT_SelectPSAMSlot(int iSlot);
int JT_GetCardType();
int JT_CloseCard();
void JT_GetCardSer(unsigned char *sCardSer);

int JT_ProReadFile(unsigned short FileID, int FileLen, unsigned char *sReply);
int JT_ProWriteFile(unsigned short FileID, int FileLen, unsigned char *sData);
int JT_ProQueryBalance(unsigned int *Balance);
int JT_ProGetCardID(unsigned char *CardID);
int JT_ProGetCardType();
int JT_ProDecrement(int iMoney,unsigned char *sTollInfo,unsigned char *sTradNo,unsigned char *sTermTradNo,char *DecTime,unsigned char *sTac);
int JT_SamGetSerialNo(unsigned char *sSerialNo);
int JT_SamGetTermID(unsigned char *sTermID);
int JT_SamReset(unsigned char *sAtr, int * iRcvLen);
int JT_AudioControl(unsigned char cTimes,unsigned char cVoice);
int JT_ReaderVersion(char* sReaderVersion);
//M1卡操作
int JT_ReadBlock(int iKeyType,int iBlockn,unsigned char *sReply);
int JT_ReadFile(unsigned short sFileID,unsigned char sKeyID,unsigned char cFileType,int iAddr,int iLength,unsigned char *sReply);
int JT_WriteBlock(int iKeyType,int iBlockn,unsigned char *sData);
int JT_WriteFile(unsigned short sFileID,unsigned char sKeyID,unsigned char cFileType,int iAddr,int iLength,unsigned char *sData);
int JT_GetCardID(char *sCardID);

#ifdef __cplusplus
}
#endif

#endif
