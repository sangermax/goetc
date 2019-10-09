#ifndef _MYEP_H_
#define _MYEP_H_

#include "pubhead.h"

#ifdef WIN32
#define API_EXPORT     __declspec(dllexport)
#elif defined(linux)
#define API_EXPORT
#endif

#ifdef __cplusplus
extern "C"
{
#endif
	API_EXPORT int Loadso();
	API_EXPORT void Closeso();

    typedef void(*pReceivedMsgCallback_t)(const char * jsonMsg);
    API_EXPORT int RegisterCallbackFunction(pReceivedMsgCallback_t callback);

	API_EXPORT void ReceiveMessages(unsigned char* dataBufIn,int dataLenIn);

	API_EXPORT void GenerateKeepaliveAckData(unsigned char** buf, int* buflen);

#ifdef __cplusplus
}
#endif		//extern "C"

#endif