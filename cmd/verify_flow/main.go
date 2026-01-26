package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const baseURL = "http://localhost:8080"

func main() {
	// Wait for server to be ready
	time.Sleep(2 * time.Second)

	fmt.Println(">>> Starting Verification Flow")

	// 1. Register
	fmt.Println("1. Registering User...")
	email := fmt.Sprintf("user-%d@example.com", time.Now().Unix())
	password := "password123"
	
	resp, err := post("/api/v1/auth/register", map[string]string{"email": email, "password": password})
	if err != nil {
		fatal("Register", err)
	}
	// no token here, need to login

	// 2. Login
	fmt.Println("2. Logging in...")
	resp, err = post("/api/v1/auth/login", map[string]string{"email": email, "password": password})
	if err != nil {
		fatal("Login", err)
	}
	token := resp["token"].(string)
	fmt.Println("   Token obtained")

	// 3. Get Deposit Address (BTC)
	fmt.Println("3. Getting Deposit Address for BTC...")
	headers := map[string]string{"Authorization": "Bearer " + token}
	body, err := get("/api/v1/deposit/address?asset=BTC", headers)
	if err != nil {
		fatal("GetAddress", err)
	}
	var addrResp map[string]string
	json.Unmarshal(body, &addrResp)
	address := addrResp["address"]
	fmt.Printf("   Address: %s\n", address)

	// 4. Simulate Deposit (10 BTC)
	fmt.Println("4. Simulating Deposit of 10 BTC (Admin)...")
	txHash := fmt.Sprintf("tx-%d", time.Now().Unix())
	_, err = post("/admin/deposit/webhook", map[string]interface{}{
		"tx_hash": txHash,
		"asset_id": "BTC",
		"address": address,
		"amount": 10.0,
	})
	if err != nil {
		fatal("SimulateDeposit", err)
	}
	fmt.Println("   Deposit Simulated")

	// 6. Request Withdrawal (5 BTC)
	fmt.Println("6. Requesting Withdrawal of 5 BTC...")
	withdrawResp, err := postAuth("/api/v1/withdraw", map[string]interface{}{
		"asset_id": "BTC",
		"amount": 5.0,
		"to_address": "external-addr-1",
	}, token)
	if err != nil {
		fatal("Withdraw", err)
	}
	fmt.Printf("   Withdrawal Requested. ID: %v\n", withdrawResp["id"])

	// 7. Process Withdrawal (Admin)
	fmt.Println("7. Processing Withdrawal Batch (Admin)...")
	_, err = post("/admin/withdrawal/process", nil)
	if err != nil {
		fatal("ProcessBatch", err)
	}
	fmt.Println("   Batch Processed")

	// 8. Get History
	fmt.Println("8. Getting History...")
	body, err = get("/api/v1/withdrawals", headers)
	if err != nil {
		fatal("GetHistory", err)
	}
	fmt.Printf("   History: %s\n", string(body))

	// 9. Generate PoR (BTC)
	fmt.Println("9. Generating PoR Snapshot...")
	porResp, err := post("/admin/por/generate?asset=BTC", nil)
	if err != nil {
		fatal("PoR", err)
	}
	fmt.Printf("   PoR Root: %v\n", porResp["merkle_root"])

	fmt.Println(">>> Verification Successful!")
}

func post(path string, data interface{}) (map[string]interface{}, error) {
	return postAuth(path, data, "")
}

func postAuth(path string, data interface{}, token string) (map[string]interface{}, error) {
	jsonData, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", baseURL+path, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	var res map[string]interface{}
	json.Unmarshal(body, &res)
	return res, nil
}

func get(path string, headers map[string]string) ([]byte, error) {
	req, _ := http.NewRequest("GET", baseURL+path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func fatal(step string, err error) {
	fmt.Printf("Failed at %s: %v\n", step, err)
	// Don't os.Exit, just panic or return
	panic(err)
}
