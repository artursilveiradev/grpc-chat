const grpc = require("@grpc/grpc-js");
const protoLoader = require("@grpc/proto-loader");
const readline = require("node:readline");

const PROTO_PATH = "../proto/chat.proto";

const packageDefinition = protoLoader.loadSync(PROTO_PATH, {
  keepCase: true,
  longs: String,
  enums: String,
  defaults: true,
  oneofs: true,
  // We must include the base directory to find timestamp.proto
  includeDirs: [
    __dirname + "/../proto",
    __dirname + "/node_modules/google-protobuf",
  ],
});

const chatProto = grpc.loadPackageDefinition(packageDefinition).chat;

const user = process.argv[2];
if (!user) {
  console.log("Usage: node client.js <username>");
  process.exit(1);
}

const GRPC_SERVER_ADDRESS = process.env.GRPC_SERVER || "localhost:50051";

const client = new chatProto.ChatService(
  GRPC_SERVER_ADDRESS,
  grpc.credentials.createInsecure()
);
console.log(`Attempting to connect to: ${GRPC_SERVER_ADDRESS}`);

const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout,
});

const call = client.Connect();

console.log(`Connected to chat as: ${user}`);

// Handle incoming messages from the server
call.on("data", (message) => {
  const ts = new Date(message.timestamp.seconds * 1000);
  const time = ts.toLocaleTimeString("en-US", {
    hour: "2-digit",
    minute: "2-digit",
  });

  // Only display messages from other users
  if (message.user !== user) {
    console.log(`\n[${time}] ${message.user}: ${message.text}`);
  }
});

// Handle the end of the stream
call.on("end", () => {
  console.log("Disconnected from server.");
  process.exit();
});

// Handle errors
call.on("error", (err) => {
  console.error("gRPC connection error:", err.details);
  process.exit(1);
});

// Send the first message to register the user
call.write({
  user: user,
  text: "Joined the room!",
});

// Read user input and send messages to the server
rl.on("line", (line) => {
  if (line.trim()) {
    call.write({
      user: user,
      text: line.trim(),
    });
  }
});

// Handle closing the terminal
rl.on("close", () => {
  console.log("Disconnecting...");
  call.end();
  process.exit(0);
});
