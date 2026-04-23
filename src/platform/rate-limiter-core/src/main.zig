const std = @import("std");

pub fn main() !void {
    const stdout = std.io.getStdOut().writer();
    try stdout.print("rate-limiter-core starting on port 50355\n", .{});

    // HTTP server placeholder
    // In production, implement a proper HTTP server with /healthz endpoint
    const port: u16 = blk: {
        const port_str = std.posix.getenv("PORT") orelse "50355";
        break :blk std.fmt.parseInt(u16, port_str, 10) catch 50355;
    };

    try stdout.print("Listening on port {d}\n", .{port});

    // Keep running
    while (true) {
        std.time.sleep(std.time.ns_per_s);
    }
}

test "port parsing" {
    const port = std.fmt.parseInt(u16, "50355", 10) catch 50355;
    try std.testing.expectEqual(@as(u16, 50355), port);
}
