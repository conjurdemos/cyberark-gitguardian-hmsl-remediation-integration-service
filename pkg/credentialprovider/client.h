#include <stdlib.h>
#include <stddef.h>
#include <stdio.h>
#include <unistd.h>
#include <string.h>
#include "cpasswordsdk.h"

/*
    Ref:
    https://docs.cyberark.com/credential-providers/10.10/en/Content/CP%20and%20ASCP/C-Application-Password-SDK.htm

    Notes:
    1. On "Ubuntu 22.04.4 LTS", copy the header and .so files to default locations, ex:
        a. copy cpasswordsdk.h to /usr/include
        b. copy libcpasswordsdk.so to /usr/lib
*/

/* CGo cannot access #define directly, so, use wrapper funcs. */
int PSDK_Get_RC_SUCCESS() { return PSDK_RC_SUCCESS; }
int PSDK_Get_RC_ERROR() { return PSDK_RC_ERROR; }
int PSDK_Get_ERROR_REQUEST_FAILED_ON_PASSWORD_CHANGE() { return PSDK_ERROR_REQUEST_FAILED_ON_PASSWORD_CHANGE; }

int ContainsError(char *msg)
{
    return strncmp("error code:", msg, 11);
}
char *ErrorCheck(ObjectHandle handle)
{
    char *msg = (char *)calloc(500, sizeof(char));
    int code;
    if (handle == NULL)
    {
        sprintf(msg, "Nil handle is invalid");
        return msg;
    }
    code = PSDK_GetErrorCode(handle);
    if (code == PSDK_RC_SUCCESS)
    {
        return NULL;
    }
    sprintf(msg, "error code: %d, error message: %s", code, PSDK_GetErrorMsg(handle));
    return msg;
}

void ReleaseHandle(ObjectHandle handle)
{
    if (handle != NULL)
    {
        PSDK_ReleaseHandle(&handle);
    }
}

char *GetPassword(ObjectHandle req, ObjectHandle *resp)
{
    int response;
    response = PSDK_GetPassword(req, resp);
    if (response == PSDK_RC_ERROR)
    {
        return ErrorCheck(req);
    }
    return NULL;
}
char *GetAttribute(ObjectHandle resp, char *name)
{
    char **pVals = NULL;
    char *retVal = NULL;
    // returns a vector of null terminated strings that are
    // filled with the values of the attribute "name"
    pVals = PSDK_GetAttribute(resp, name);
    if (pVals == NULL)
    {
        return ErrorCheck(resp);
    }
    retVal = (char *)calloc(strlen(pVals[0]) + 1, sizeof(char));
    strcpy(retVal, pVals[0]);

    PSDK_ReleaseAttributeData(&pVals);

    return retVal;
}
