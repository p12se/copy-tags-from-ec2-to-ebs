resource "aws_cloudwatch_event_rule" "trigger_lambda_12_hours" {
  name        = "trigger_lambda_12_hours"
  description = "trigger_lambda_12_hours"
  schedule_expression = "rate(12 hours)"
  depends_on = [
    "aws_lambda_function.copy-tags-from-ec2-to-ebs"
  ]

}
resource "aws_lambda_permission" "allow_cloudwatch_to_call_copy-tags-from-ec2-to-ebs" {
    statement_id = "AllowExecutionFromCloudWatch"
    action = "lambda:InvokeFunction"
    function_name = "${aws_lambda_function.copy-tags-from-ec2-to-ebs.function_name}"
    principal = "events.amazonaws.com"
    source_arn = "${aws_cloudwatch_event_rule.trigger_lambda_12_hours.arn}"
}

resource "aws_cloudwatch_event_target" "copy-tags-from-ec2-to-ebs_target" {
  target_id = "copy-tags-from-ec2-to-ebs" // Worked for me after I added `target_id`
  rule = "${aws_cloudwatch_event_rule.trigger_lambda_12_hours.name}"
  arn = "${aws_lambda_function.copy-tags-from-ec2-to-ebs.arn}"
}


resource "aws_iam_role_policy" "iam_policy_for_copy_tags_lambda" {
  name = "iam_policy_for_copy_tags_lambda"
  role = "${aws_iam_role.iam_role_for_copy_tags_lambda.id}"

  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "VisualEditor0",
            "Effect": "Allow",
            "Action": "ec2:CreateTags",
            "Resource": [
                "arn:aws:ec2:*:*:instance/*",
                "arn:aws:ec2:*:*:volume/*"
            ]
        },
        {
            "Sid": "VisualEditor1",
            "Effect": "Allow",
            "Action": [
                "ec2:DescribeInstances",
                "ec2:DescribeVolumes"
            ],
            "Resource": "*"
        }
    ]
}
EOF
}

resource "aws_iam_role" "iam_role_for_copy_tags_lambda" {
  name = "iam_role_for_copy_tags_lambda"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}


resource "aws_lambda_function" "copy-tags-from-ec2-to-ebs" {
  filename         = "../main.zip"
  function_name    = "copy-tags-from-ec2-to-ebs"
  role             = "${aws_iam_role.iam_role_for_copy_tags_lambda.arn}"
  handler          = "main"
  source_code_hash = "${base64sha256(file("../main.zip"))}"
  runtime          = "go1.x"
}
