namespace Cim4Examples;

// 1. Tasks and jobs: one request describes the whole pipeline.
// A job is a set of named tasks. A task reads the output of another task through "input".
public static class TasksAndJobs
{
    public static async Task Run()
    {
        var job = (await Api.Call(HttpMethod.Post, "/v1/jobs", new
        {
            tag = "example-01",
            tasks = new Dictionary<string, object>
            {
                ["import-logo"] = new { operation = "import/url", url = Api.SampleUrl },
                ["to-jpg"] = new { operation = "convert", input = "import-logo", output_format = "jpg", options = new { background = "#ffffff" } },
                ["export"] = new { operation = "export/url", input = "to-jpg" },
            },
        }))!;
        Console.WriteLine($"Created job {job["id"]} status {job["status"]}");
        Api.PrintTasks(job);

        // Block until the job settles (90 s per call, repeated). Example 02 polls instead, example 03 listens.
        var done = await Api.WaitForJob(job["id"]!.GetValue<string>());
        Console.WriteLine($"Job {done["id"]} {done["status"]}");
        Api.PrintTasks(done);
        Console.WriteLine($"Download: {done["tasks"]!["export"]!["result"]?["url"]}");
    }
}
