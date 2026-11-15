namespace Cim4Examples;

// 6. A complex job in one request:
//   page 1: a capture of a web page as PDF
//   page 2: an image converted to PDF
//   merge both into one PDF, make a 200 px thumbnail of the image,
//   pack the PDF and the thumbnail into a zip and export it.
public static class ComplexJob
{
    public static async Task Run()
    {
        var job = (await Api.Call(HttpMethod.Post, "/v1/jobs", new
        {
            tag = "example-06",
            tasks = new Dictionary<string, object>
            {
                ["page"] = new { operation = "capture-website", url = "https://convertintomp4.com/api-docs", output_format = "pdf", options = new { page_format = "A4", print_background = true } },
                ["logo"] = new { operation = "import/url", url = Api.SampleUrl },
                ["logo-pdf"] = new { operation = "convert", input = "logo", output_format = "pdf" },
                ["merged"] = new { operation = "merge", input = new[] { "page", "logo-pdf" }, output_format = "pdf" },
                ["thumb"] = new { operation = "thumbnail", input = "logo", output_format = "jpg", options = new { width = 200, height = 200, fit = "max" } },
                ["package"] = new { operation = "archive/create", input = new[] { "merged", "thumb" }, output_format = "zip" },
                ["export"] = new { operation = "export/url", input = "package" },
            },
        }))!;
        Console.WriteLine($"Created job {job["id"]} with {job["tasks"]!.AsObject().Count} tasks");

        var done = await Api.WaitForJob(job["id"]!.GetValue<string>());
        Console.WriteLine($"Job {done["status"]}");
        Api.PrintTasks(done);
        if (done["status"]!.GetValue<string>() == "completed")
            Console.WriteLine($"Zip: {done["tasks"]!["export"]!["result"]!["url"]}");
    }
}
