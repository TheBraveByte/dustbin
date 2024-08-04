#include <SPI.h>
#include <MFRC522.h>
#include <WiFi.h>
#include <HTTPClient.h>

// RFID Reader Pins
#define RST_PIN 9
#define SS_PIN 10

MFRC522 mfrc522(SS_PIN, RST_PIN); // Create MFRC522 instance

// Wi-Fi credentials
const char* ssid = "your_SSID";
const char* password = "your_PASSWORD";

// Server URL
const char* serverUrl = "http://<your_server_ip>:5000/rfid";

// Initialize Wi-Fi and HTTP client
WiFiClient client;
HTTPClient http;

void setup() {
    Serial.begin(9600); // Initialize serial communications with the PC
    SPI.begin();        // Initialize SPI bus
    mfrc522.PCD_Init(); // Initialize RFID reader

    // Connect to Wi-Fi
    Serial.print("Connecting to ");
    Serial.println(ssid);
    WiFi.begin(ssid, password);

    while (WiFi.status() != WL_CONNECTED) {
        delay(500);
        Serial.print(".");
    }

    Serial.println("Connected to Wi-Fi");
}
mongodb+srv://ayaaakinleye:BYObCPGZExCPQceD@cluster0.opv1wfb.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0")
void loop() {
    // Check if a new RFID card is present
    if (mfrc522.PICC_IsNewCardPresent() && mfrc522.PICC_ReadCardSerial()) {
        String rfid = "";
        for (byte i = 0; i < mfrc522.uid.size; i++) {
            rfid += String(mfrc522.uid.uidByte[i] < 0x10 ? "0" : "");
            rfid += String(mfrc522.uid.uidByte[i], HEX);
        }
        rfid.toUpperCase();

        Serial.print("RFID Tag Detected: ");
        Serial.println(rfid);

        // Send RFID tag data to the server
        String response = sendRfidToServer(rfid);
        
        // Handle server response
        if (response == "Waste bin opened and charge processed") {
            // Open waste bin
            openWasteBin();
            Serial.println("Bin opened. Charge amount processed.");
            delay(5000); // Keep bin open for 5 seconds
            closeWasteBin(); // Close the bin after 5 seconds
        } else {
            // Alert the user (e.g., via buzzer or LED)
            alertUser(response);
        }

        // Halt PICC to prevent additional reads
        mfrc522.PICC_HaltA();
        mfrc522.PCD_StopCrypto1();
    }
}

String sendRfidToServer(String rfid) {
    if (WiFi.status() == WL_CONNECTED) {
        http.begin(client, serverUrl);
        http.addHeader("Content-Type", "application/json");

        // Prepare the JSON payload
        String payload = "{\"tag_id\":\"" + rfid + "\"}";

        int httpCode = http.POST(payload);
        String response = "";

        if (httpCode > 0) {
            response = http.getString();
            Serial.println("Server Response: " + response);
        } else {
            Serial.println("Error sending data to server");
            response = "error";
        }

        http.end();
        return response;
    } else {
        Serial.println("Wi-Fi not connected");
        return "error";
    }
}

void openWasteBin() {
    // Example: Set pin HIGH to activate relay (open bin)
    pinMode(8, OUTPUT);
    digitalWrite(8, HIGH);
}

void closeWasteBin() {
    // Example: Set pin LOW to deactivate relay (close bin)
    digitalWrite(8, LOW);
}

void alertUser(String message) {
    // Example: Use an LED or buzzer to alert the user
    pinMode(7, OUTPUT);
    digitalWrite(7, HIGH); // Turn on alert (LED or buzzer)
    delay(2000); // Alert for 2 seconds
    digitalWrite(7, LOW); // Turn off alert
}
