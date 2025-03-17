// SPDX-License-Identifier: GPL-3.0

pragma solidity >=0.7.0 <0.9.0;

contract SvmSetter {
    struct InstructionAccount {
        uint32 id_index;       // id_index
        uint32 caller_index;   // caller_index
        uint32 callee_index;   // callee_index
        bool is_signer;        // is_signer
        bool is_writable;      // is_writable
    }

    // Struct for Instruction
    struct Instruction {
        uint32[] program_index;       
        InstructionAccount[] accounts;
        bytes data;
    }

    // Struct for MsgTransaction
    struct MsgTransaction {
        string[] signers;          // signers (repeated string)
        string[] accounts;         // accounts (repeated string)
        Instruction[] instructions; // instructions (repeated Instruction)
        uint64 compute_budget;      // compute_budget (uint64)
    }

    string public evmString;

    function set(string memory _evmString) public {
        evmString = _evmString;
    }

    // TODO: Temporarily leave the svm fee payer here, we should let contract be a fee payer
    function setSvm(string memory svmFeePayer, string memory _svmString) public returns (bytes memory result) {
        address svmProxy = address(0xaB6B4d064C968eCa87F775d2493a222987052BC0);
        MsgTransaction memory transaction;
        transaction.accounts = new string[](4);
        transaction.accounts[0] = "83zfZYacFrGq5eBnnp6EQPxapcpjpxdjAKpLavqtSJ32";
        transaction.accounts[1] = "22z6H6DeW4ULByQKLERtHjSoDeFiuCii8v3NuLSAf2db";
        transaction.accounts[2] = svmFeePayer;
        transaction.accounts[3] = "11111111111111111111111111111111";
        transaction.compute_budget = 1000000;

        // build instruction
        Instruction memory instruction;
        instruction.program_index = new uint32[](1);
        instruction.program_index[0] = 0;
        instruction.accounts = new InstructionAccount[](3);
        instruction.accounts[0] = InstructionAccount({
            id_index: 1,
            caller_index: 1,
            callee_index: 0,
            is_signer: false,
            is_writable: true
        });
        instruction.accounts[1] = InstructionAccount({
            id_index: 2,
            caller_index: 2,
            callee_index: 1,
            is_signer: true,
            is_writable: true
        });
        instruction.accounts[2] = InstructionAccount({
            id_index: 3,
            caller_index: 3,
            callee_index: 2,
            is_signer: false,
            is_writable: false
        });
        bytes memory discriminator = hex"df725b88c54e9999";
        bytes4 stringLen = this.getStringLengthLE(_svmString);
        instruction.data = bytes.concat(discriminator, stringLen, bytes(_svmString));
        transaction.instructions = new Instruction[](1);
        transaction.instructions[0] = instruction;

        (bool ok, bytes memory r) = svmProxy.call(abi.encode(transaction));
        require(ok, "call must succeed");
        return r;
    }

    function conditionalSetSvm(string memory svmFeePayer) public returns (bytes memory result) {
        if (keccak256(bytes(getSvm())) == keccak256("svm")) {
            return setSvm(svmFeePayer, "evm");
        } 
        return setSvm(svmFeePayer, "svm");
    }

    // substring for `memory` does not natively supported by solidity, use for loop as example now
    function parseSvmResponse(bytes memory data) internal pure returns (string memory) {
        bytes memory res = new bytes(data.length - 4);
        for(uint256 i = 4; i<data.length; i++) {
            res[i-4] = data[i];
        }

        return string(res);
    }

    function getSvm() public view returns (string memory) {
        address svmProxy = address(0xaB6B4d064C968eCa87F775d2493a222987052BC0);
        MsgTransaction memory transaction;
        transaction.accounts = new string[](2);
        transaction.accounts[0] = "83zfZYacFrGq5eBnnp6EQPxapcpjpxdjAKpLavqtSJ32";
        transaction.accounts[1] = "22z6H6DeW4ULByQKLERtHjSoDeFiuCii8v3NuLSAf2db";
        transaction.compute_budget = 1000000;
         Instruction memory instruction;
        instruction.program_index = new uint32[](1);
        instruction.program_index[0] = 0;
        instruction.accounts = new InstructionAccount[](1);
        instruction.accounts[0] = InstructionAccount({
            id_index: 1,
            caller_index: 1,
            callee_index: 0,
            is_signer: false,
            is_writable: false
        });
        bytes memory discriminator = hex"dc8bfd5f8098939f";
        instruction.data = discriminator;
        transaction.instructions = new Instruction[](1);
        transaction.instructions[0] = instruction;
        (bool ok, bytes memory r) = svmProxy.staticcall(abi.encode(transaction));
        require(ok, "call must succeed");

        return parseSvmResponse(r);
    }

    function getStringLengthLE(string memory str) public pure returns (bytes4) {
        uint256 len = bytes(str).length;
        return toLittleEndian(len);
    }

    function toLittleEndian(uint256 value) public pure returns (bytes4 result) {
        // Extract last 4 bytes
        bytes32 last4 = bytes32(value);
        // Reverse the bytes
        return bytes4(
            (uint32(uint8(last4[31])) << 24) |
            (uint32(uint8(last4[30])) << 16) |
            (uint32(uint8(last4[29])) << 8) |
            (uint32(uint8(last4[28])))
        );
    }
}
