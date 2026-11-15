// Usage: dotnet run -- <example>   (01 to 06)
using Cim4Examples;

if (string.IsNullOrEmpty(Api.Key))
{
    Console.Error.WriteLine("Set CIM4_API_KEY to an API key from https://convertintomp4.com/account/api/keys");
    return 1;
}
var examples = new Dictionary<string, Func<Task>>
{
    ["01"] = TasksAndJobs.Run,
    ["02"] = CheckStatusWithPolling.Run,
    ["03"] = RealtimeWebsocketUpdates.Run,
    ["04"] = UploadingFiles.Run,
    ["05"] = HandlingResultsAndErrors.Run,
    ["06"] = ComplexJob.Run,
};
if (args.Length < 1 || !examples.TryGetValue(args[0], out var run))
{
    Console.Error.WriteLine("usage: dotnet run -- 01|02|03|04|05|06");
    return 2;
}
await run();
return 0;
