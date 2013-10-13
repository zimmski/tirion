#ifndef tirion_h_INCLUDED
#define tirion_h_INCLUDED

/**
 * @file   tirion.h
 * @brief  The public API of Tirion's C client library
 */

#include <stdbool.h>

/**
 * The version of the tirion client
 * The version is also used in the communication with the agent and
 * dictates the whole communication protocol.
 */
#define TIRION_VERSION "0.1"

/**
 * Error codes for all Tirion functions
 */
enum {
	TIRION_OK,                            /**< everything is ok */
	TIRION_ERROR_METRIC_COUNT,            /**< did not receive a correct metric count */
	TIRION_ERROR_SET_SID,                 /**< could not create a new process session and group id */
	TIRION_ERROR_SHM_ATTACH,              /**< could not attach the shm */
	TIRION_ERROR_SHM_DETACH,              /**< could not detach the shm */
	TIRION_ERROR_SHM_KEY,                 /**< could not create a shm key */
	TIRION_ERROR_SHM_INIT,                /**< could not initialized the shm */
	TIRION_ERROR_SHM_PATH,                /**< did not receive a correct shm path */
	TIRION_ERROR_SHM_NO_INIT,             /**< the shm cannot be used uninitialized */
	TIRION_ERROR_SOCKET_CONNECT,          /**< the socket for the agent communication could not connect */
	TIRION_ERROR_SOCKET_CREATE,           /**< the socket for the agent communication could not be created */
	TIRION_ERROR_SOCKET_RECEIVE,          /**< could not receive a message over the socket */
	TIRION_ERROR_SOCKET_SHUTDOWN,         /**< could not shut down the socket for the agent communication */
	TIRION_ERROR_SOCKET_SEND,             /**< could not send a message over the socket */
	TIRION_ERROR_THREAD_HANDLE_COMMANDS,  /**< could not create the handle commands pthread */
	TIRION_ERROR_THREAD_JOIN,             /**< could not join a pthread with the parent process */
};

/**
 * Struct for private data of a Tirion object
 * Tirion uses the pimpl idiom (opaque pointer) to hide the internal state
 * from the user. This allows many independent objects without interference.
 */
typedef struct TirionPrivateStruct TirionPrivate;

/**
 * Struct for a Tirion object
 */
typedef struct TirionStruct {
	bool running;       /**< States if the object is running */
	bool verbose;       /**< States if the object produces verbose output */
	TirionPrivate *p;   /**< private data of the Tirion object, DO NOT TOUCH */
} Tirion;

/**
 * Create a new Tirion object
 *
 * @param socket the socket filepath to connect to the agent
 * @param verbose enable or disable verbose output of the client library
 *
 * @return a Tirion object on success or null if the allocation failed
 */
Tirion *tirionNew(const char *socket, bool verbose);

/**
 * Initialize a Tirion object
 *
 * @param tirion the Tirion object to be initialized
 *
 * @return error code
 *    - TIRION_OK everything is ok
 *    - TIRION_ERROR_METRIC_COUNT did not receive a correct metric count
 *    - TIRION_ERROR_SET_SID could not create a new process session and group id
 *    - TIRION_ERROR_SHM_ATTACH could not attach the shm
 *    - TIRION_ERROR_SHM_KEY could not create a shm key
 *    - TIRION_ERROR_SHM_INIT could not initialized the shm
 *    - TIRION_ERROR_SHM_NO_INIT the shm cannot be used uninitialized
 *    - TIRION_ERROR_SHM_PATH did not receive a correct shm pathx
 *    - TIRION_ERROR_SOCKET_CONNECT the socket for the agent communication could not connect
 *    - TIRION_ERROR_SOCKET_CREATE the socket for the agent communication could not be created
 *    - TIRION_ERROR_SOCKET_RECEIVE could not receive a message over the socket
 *    - TIRION_ERROR_SOCKET_SEND could not send a message over the socket
 *    - TIRION_ERROR_THREAD_HANDLE_COMMANDS could not create the handle commands pthread
 */
long tirionInit(Tirion *tirion);

/**
 * Uninitialized a Tirion object
 *
 * @param tirion the Tirion object to be uninitialized
 *
 * @return error code
 *    - TIRION_OK everything is ok
 *    - TIRION_ERROR_SHM_DETACH could not detach the shm
 *    - TIRION_ERROR_SOCKET_SHUTDOWN could not shut down the socket for the agent communication
 *    - TIRION_ERROR_THREAD_JOIN could not join a pthread with the parent process
 */
long tirionClose(Tirion *tirion);

/**
 * Destroy a Tirion object
 *
 * @param tirion the Tirion object to be destroyed
 *
 * @return error code
 *    - TIRION_OK everything is ok
 */
long tirionDestroy(Tirion *tirion);

/**
 * Add a value to a metric
 *
 * @param tirion the Tirion object
 * @param i the index of the metric
 * @param v the value to be add to the metric
 *
 * @return the new value of the metric
 */
float tirionAdd(Tirion *tirion, long i, float v);

/**
 * Decrement a metric by 1.0
 *
 * @param tirion the Tirion object
 * @param i the index of the metric
 *
 * @return the new value of the metric
 */
float tirionDec(Tirion *tirion, long i);

/**
 * Increment a metric by 1.0
 *
 * @param tirion the Tirion object
 * @param i the index of the metric
 *
 * @return the new value of the metric
 */
float tirionInc(Tirion *tirion, long i);

/**
 * Subtract a value of a metric
 *
 * @param tirion the Tirion object
 * @param i the index of the metric
 * @param v the value to be subtracted of the metric
 *
 * @return the new value of the metric
 */
float tirionSub(Tirion *tirion, long i, float v);

/**
 * Send a tag to the socket
 *
 * @param tirion the Tirion object
 * @param format the tag string that follows the same specifications as format in printf
 * @param ... additional arguments for format
 *
 * @return error code
 *    - TIRION_OK everything is ok
 *    - TIRION_ERROR_SOCKET_SEND could not send a message over the socket
 */
long tirionTag(Tirion *tirion, const char *format, ...);

/**
 * Output a Tirion debug message
 *
 * @param tirion the Tirion object
 * @param format the message string that follows the same specifications as format in printf
 * @param ... additional arguments for format
 */
void tirionD(const Tirion *tirion, const char *format, ...);

/**
 * Output a Tirion error message
 *
 * @param tirion the Tirion object
 * @param format the message string that follows the same specifications as format in printf
 * @param ... additional arguments for format
 */
void tirionE(const Tirion *tirion, const char *format, ...);

/**
 * Output a Tirion verbose message
 *
 * @param tirion the Tirion object
 * @param format the message string that follows the same specifications as format in printf
 * @param ... additional arguments for format
 */
void tirionV(const Tirion *tirion, const char *format, ...);

#endif
