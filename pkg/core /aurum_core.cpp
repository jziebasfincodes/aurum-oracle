// aurum_core.cpp - Reference Implementation for Hardened Storage Daemon
// Compile: g++ -std=c++17 aurum_core.cpp -o aurum_core -lssl -lcrypto

#include <iostream>
#include <fstream>
#include <vector>
#include <string>
#include <ctime>
#include <sstream>
#include <iomanip>
#include <openssl/sha.h>
#include <openssl/evp.h>

using namespace std;

// --- Structures ---

struct Transaction {
    string tx_hash;
    long timestamp;
    string data_json;
};

struct Block {
    long index;
    long timestamp;
    string prev_hash;
    string merkle_root;
    string hash;
    string signature;
    vector<Transaction> transactions;
};

// --- Crypto Utils ---

string sha256(const string str) {
    unsigned char hash[SHA256_DIGEST_LENGTH];
    SHA256_CTX sha256;
    SHA256_Init(&sha256);
    SHA256_Update(&sha256, str.c_str(), str.size());
    SHA256_Final(hash, &sha256);
    
    stringstream ss;
    for(int i = 0; i < SHA256_DIGEST_LENGTH; i++) {
        ss << hex << setw(2) << setfill('0') << (int)hash[i];
    }
    return ss.str();
}

string compute_merkle_root(const vector<Transaction>& txs) {
    if (txs.empty()) return sha256("empty");
    
    vector<string> hashes;
    for (const auto& tx : txs) hashes.push_back(tx.tx_hash);
    
    while (hashes.size() > 1) {
        vector<string> next_level;
        for (size_t i = 0; i < hashes.size(); i += 2) {
            if (i + 1 < hashes.size()) {
                next_level.push_back(sha256(hashes[i] + hashes[i+1]));
            } else {
                next_level.push_back(hashes[i]);
            }
        }
        hashes = next_level;
    }
    return hashes[0];
}

// --- Storage Engine ---

class Ledger {
    string filename;
public:
    Ledger(string f) : filename(f) {}

    bool append_block(Block& b) {
        // 1. Calculate Merkle
        b.merkle_root = compute_merkle_root(b.transactions);
        
        // 2. Calculate Hash
        stringstream ss;
        ss << b.index << b.timestamp << b.prev_hash << b.merkle_root;
        b.hash = sha256(ss.str());
        
        // 3. Write to disk (Binary append)
        ofstream outfile(filename, ios::binary | ios::app);
        if (!outfile.is_open()) return false;

        // Simple serialization format:
        // [Index 8B][Time 8B][HashLen 4B][HashBytes][...]
        outfile.write(reinterpret_cast<char*>(&b.index), sizeof(b.index));
        outfile.write(reinterpret_cast<char*>(&b.timestamp), sizeof(b.timestamp));
        
        size_t hash_len = b.hash.size();
        outfile.write(reinterpret_cast<char*>(&hash_len), sizeof(int));
        outfile.write(b.hash.c_str(), hash_len);
        
        outfile.close();
        
        cout << "Locked Block #" << b.index << " Hash: " << b.hash << endl;
        return true;
    }
};

// --- Daemon Entry ---

int main(int argc, char* argv[]) {
    cout << "AURUM C++ Core Daemon v1.0" << endl;
    cout << "Initializing Secure Storage..." << endl;
    
    Ledger ledger("aurum_ledger.dat");
    
    // Simulating a block add (In prod, this listens on a socket)
    Block b;
    b.index = 1;
    b.timestamp = std::time(nullptr);
    b.prev_hash = "00000000000000000000000000000000";
    
    Transaction tx;
    tx.tx_hash = sha256("price:2000");
    b.transactions.push_back(tx);
    
    if(ledger.append_block(b)) {
        cout << "Success: Block committed." << endl;
    } else {
        cerr << "Error: Storage failure." << endl;
    }
    
    return 0;
}