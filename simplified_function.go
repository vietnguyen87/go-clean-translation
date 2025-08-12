func test1(req *bo.ReservationFlightOrderRequest, mapListOrderDetail map[constant.FlightDirectionType][]*models.FlightOrderDetail) {
	// Extract departure and return order details
	departureDetail := req.OrderDetails[0]
	returnDetail := req.OrderDetails[1]
	
	// Get departure order details from the map
	departureOrderDetails := mapListOrderDetail[constant.FlightOrderDetailTypeDeparture]
	
	// Create a new map for departure details only
	departureMap := map[constant.FlightDirectionType][]*models.FlightOrderDetail{
		constant.FlightOrderDetailTypeDeparture: departureOrderDetails,
	}
	
	// Create a copy of the request with only departure details
	departureReq := *req
	departureReq.OrderDetails = []*bo.FlightReservationDetail{departureDetail}
	
	// Now you can use departureReq and departureMap for departure processing
	// ... continue with your logic
}