package handlers

import (
	"general/convert"
	"general/dberror"
	"general/server"
	"general/types"
	"net/http"
)

// GetLikes get the likes from an user bounded by the given offset and max in the request. The results are ordered by name of the song
func (handler *LikesHandler) GetLikes(response http.ResponseWriter, request *http.Request) {
	user := request.Context().Value(types.Credentials{}).(types.Credentials)
	offsetMax := request.Context().Value(server.OffsetMax{}).(server.OffsetMax)
	offset, max := offsetMax.Offset, offsetMax.Max
	handler.Logger.Printf("Received call for likes of user %v and limit %v,%v\n", user.Username, offset, max)
	results, errorSearch := handler.db.GetLikes(user.ID, offset, max+1)
	if errorSearch != nil {
		errorcode := errorSearch.(dberror.DBError).ErrorCode
		if errorcode == dberror.InvalidOffsetMax {
			badRequests.Inc()
			handler.Logger.Printf("Request with invalid  values for query parameters: %v,%v", offset, max)
			server.SendError(response, http.StatusBadRequest)
			return
		}
		if errorcode == dberror.NotFoundError {
			handler.Logger.Printf("Request with no results for user %v: %v,%v", user.Username, offset, max)
			server.SendError(response, http.StatusNotFound)
			return
		}
		failureGetRequest.Inc()
		handler.Logger.Printf("[Error] Can't find likes of user %v and limit %v,%v due to: %s\n", user.Username, offset, max, errorSearch)
		server.SendError(response, http.StatusInternalServerError)
		return
	}
	if len(results) == 0 {
		handler.Logger.Printf("Failed to find likes of user %v and limit %v,%v\n", user.Username, offset, max)
		failureGetRequest.Inc()
		server.SendError(response, http.StatusNotFound)
		return
	}
	handler.Logger.Printf("Succesfully found %v likes of user %v and limit %v,%v\n", len(results), user.Username, offset, max)
	hasNext := (len(results) > max)
	if hasNext {
		results = results[0:max]
	}
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	err := convert.WriteToJSON(&types.MultipleSongs{Data: results, HasNext: hasNext}, response)
	if err != nil {
		handler.Logger.Printf("[ERROR] %s\n", err)
	}
}

// GetDislikes get the dislikes from an user bounded by the given offset and max in the request. The results are ordered by name of the song
func (handler *LikesHandler) GetDislikes(response http.ResponseWriter, request *http.Request) {
	user := request.Context().Value(types.Credentials{}).(types.Credentials)
	offsetMax := request.Context().Value(server.OffsetMax{}).(server.OffsetMax)
	offset, max := offsetMax.Offset, offsetMax.Max
	handler.Logger.Printf("Received call for dislikes of user %v and limit %v,%v\n", user.Username, offset, max)
	results, errorSearch := handler.db.GetDislikes(user.ID, offset, max+1)
	if errorSearch != nil {
		errorcode := errorSearch.(dberror.DBError).ErrorCode
		if errorcode == dberror.InvalidOffsetMax {
			badRequests.Inc()
			handler.Logger.Printf("Request with invalid  values for query parameters: %v,%v", offset, max)
			server.SendError(response, http.StatusBadRequest)
			return
		}
		if errorcode == dberror.NotFoundError {
			handler.Logger.Printf("Request with no results for user %v: %v,%v", user.Username, offset, max)
			server.SendError(response, http.StatusNotFound)
			return
		}
		failureGetRequest.Inc()
		handler.Logger.Printf("[Error] Can't find dislikes of user %v and limit %v,%v due to: %s\n", user.Username, offset, max, errorSearch)
		server.SendError(response, http.StatusInternalServerError)
		return
	}
	if len(results) == 0 {
		handler.Logger.Printf("Failed to find dislikes of user %v and limit %v,%v\n", user.Username, offset, max)
		failureGetRequest.Inc()
		server.SendError(response, http.StatusNotFound)
		return
	}
	handler.Logger.Printf("Succesfully found %v dislikes of user %v and limit %v,%v\n", len(results), user.Username, offset, max)
	hasNext := (len(results) > max)
	if hasNext {
		results = results[0:max]
	}
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	err := convert.WriteToJSON(&types.MultipleSongs{Data: results, HasNext: hasNext}, response)
	if err != nil {
		handler.Logger.Printf("[ERROR] %s\n", err)
	}
}
