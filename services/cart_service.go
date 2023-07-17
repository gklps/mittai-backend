package services

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gklps/mittai-backend/db"
	"github.com/gklps/mittai-backend/models"
	"github.com/gorilla/mux"
)

// CartService represents a service for cart-related operations
type CartService struct {
	Repository *db.Repository
	CartDB     map[int]*models.Cart
}

// NewCartService creates a new CartService instance
func NewCartService(repository *db.Repository) *CartService {
	return &CartService{
		Repository: repository,
		CartDB:     make(map[int]*models.Cart),
	}
}

// RegisterRoutes registers the cart service routes
func (cs *CartService) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/cart/{userID}", cs.AddToCart).Methods("POST")
	r.HandleFunc("/cart/{userID}", cs.GetCartByUserID).Methods("GET")
	r.HandleFunc("/cart/{userID}/{productWeightID}", cs.UpdateCartItem).Methods("PUT")
	r.HandleFunc("/cart/{userID}/{productWeightID}", cs.RemoveCartItem).Methods("DELETE")
}

// AddToCart adds an item to the user's cart
// @Summary Add an item to the cart
// @Tags Cart
// @Accept json
// @Produce json
// @Param userID path string true "User ID"
// @Param cartItem body models.CartItem true "Cart item object"
// @Success 200 {string} string "Item added to cart successfully"
// @Failure 400 {object} ErrorResponse "Invalid request body"
// @Router /cart/{userID} [post]
func (cs *CartService) AddToCart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var cartItem models.CartItem
	err := json.NewDecoder(r.Body).Decode(&cartItem)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userIDstr := vars["userID"]

	userID, err := strconv.Atoi(userIDstr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Get the user's cart from the database
	cart, exists := cs.CartDB[userID]
	if !exists {
		cart = &models.Cart{
			UserID:    userID,
			Items:     make([]*models.CartItem, 0),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		cs.CartDB[userID] = cart
	}

	// Check if the item already exists in the cart
	for _, item := range cart.Items {
		if item.Product.ID == cartItem.Product.ID {
			// Item already exists, update the quantity
			item.Quantity += cartItem.Quantity
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Item added to cart successfully"))
			return
		}
	}

	// Item does not exist, create a new cart item

	// Fetch the product weight details from the database based on productWeightID
	productWeightID := cartItem.Product.ID
	productWeight, err := cs.getProductWeightByID(productWeightID)
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to fetch product weight details", http.StatusInternalServerError)
		return
	}

	// Create a new cart item with the fetched product weight details
	newCartItem := &models.CartItem{
		Product:   productWeight,
		Quantity:  cartItem.Quantity,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Add the new cart item to the cart
	cart.Items = append(cart.Items, newCartItem)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Item added to cart successfully"))
}

// GetCartByUserID retrieves the user's cart by user ID
// @Summary Get the user's cart by user ID
// @Tags Cart
// @Accept json
// @Produce json
// @Param userID path string true "User ID"
// @Success 200 {object} models.Cart "User's cart retrieved successfully"
// @Router /cart/{userID} [get]
func (cs *CartService) GetCartByUserID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDstr := vars["userID"]

	userID, err := strconv.Atoi(userIDstr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	cart, exists := cs.CartDB[userID]
	if !exists {
		cart = &models.Cart{
			UserID:    userID,
			Items:     make([]*models.CartItem, 0),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		cs.CartDB[userID] = cart
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cart)
}

// UpdateCartItem updates the quantity of a cart item in the user's cart
// @Summary Update the quantity of a cart item
// @Tags Cart
// @Accept json
// @Produce json
// @Param userID path string true "User ID"
// @Param productWeightID path string true "Product Weight ID"
// @Param cartItem body models.CartItem true "Cart item object"
// @Success 200 {string} string "Cart item updated successfully"
// @Failure 400 {object} ErrorResponse "Invalid request body"
// @Failure 404 {object} ErrorResponse "Cart item not found"
// @Router /cart/{userID}/{productWeightID} [put]
func (cs *CartService) UpdateCartItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productWeightIDStr := vars["productWeightID"]

	var cartItem models.CartItem
	err := json.NewDecoder(r.Body).Decode(&cartItem)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userIDstr := vars["userID"]

	userID, err := strconv.Atoi(userIDstr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Get the user's cart from the database
	cart, exists := cs.CartDB[userID]
	if !exists {
		http.Error(w, "Cart not found", http.StatusNotFound)
		return
	}

	productWeightID, err := strconv.Atoi(productWeightIDStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid product weight ID", http.StatusBadRequest)
		return
	}

	// Find the cart item by matching the product's ID
	for _, item := range cart.Items {
		if item.Product != nil && item.Product.ID == productWeightID {
			// Fetch the product weight details from the database based on productWeightID
			productWeight, err := cs.getProductWeightByID(productWeightID)
			if err != nil {
				log.Println(err)
				http.Error(w, "Failed to fetch product weight details", http.StatusInternalServerError)
				return
			}

			// Update the cart item with the fetched product weight details
			item.Product = productWeight
			item.Quantity = cartItem.Quantity

			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Cart item updated successfully"))
			return
		}
	}

	http.Error(w, "Cart item not found", http.StatusNotFound)
}

// RemoveCartItem removes a cart item from the user's cart
// @Summary Remove a cart item from the cart
// @Tags Cart
// @Param userID path string true "User ID"
// @Param productWeightID path string true "Product Weight ID"
// @Success 200 {string} string "Cart item removed successfully"
// @Failure 404 {object} ErrorResponse "Cart item not found"
// @Router /cart/{userID}/{productWeightID} [delete]
func (cs *CartService) RemoveCartItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productWeightIDStr := vars["productWeightID"]
	userIDstr := vars["userID"]

	userID, err := strconv.Atoi(userIDstr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Get the user's cart from the database
	cart, exists := cs.CartDB[userID]
	if !exists {
		http.Error(w, "Cart not found", http.StatusNotFound)
		return
	}

	productWeightID, err := strconv.Atoi(productWeightIDStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid product weight ID", http.StatusBadRequest)
		return
	}

	// Find the index of the cart item by matching the product's ID
	index := -1
	for i, item := range cart.Items {
		if item.Product.ID == productWeightID {
			index = i
			break
		}
	}

	if index == -1 {
		http.Error(w, "Cart item not found", http.StatusNotFound)
		return
	}

	// Remove the cart item from the cart
	cart.Items = append(cart.Items[:index], cart.Items[index+1:]...)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Cart item removed successfully"))
}

// getProductWeightByID fetches the product weight details from the database based on the given productWeightID
func (cs *CartService) getProductWeightByID(productWeightID int) (*models.ProductWeight, error) {
	query := "SELECT id, product_id, weight, price, stock_availability, created_at, updated_at FROM product_weights WHERE id = ?"

	// Prepare the query
	stmt, err := cs.Repository.DB.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	// Execute the query
	row := stmt.QueryRow(productWeightID)

	// Create variables to store the retrieved data
	var (
		id                int
		productID         int
		weight            int
		price             float64
		stockAvailability int
		createdAt         time.Time
		updatedAt         time.Time
	)

	// Scan the row data into the variables
	err = row.Scan(&id, &productID, &weight, &price, &stockAvailability, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	// Create and return the product weight instance
	productWeight := &models.ProductWeight{
		ID:                id,
		ProductID:         productID,
		Weight:            weight,
		Price:             price,
		StockAvailability: stockAvailability,
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
	}

	return productWeight, nil
}