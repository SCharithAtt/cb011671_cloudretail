# ─────────────────────────────────────────────────────────────────────────────
# EC2 Auto Scaling for ECS Cluster
# ─────────────────────────────────────────────────────────────────────────────

# ── Latest ECS-Optimized AMI ────────────────────────────────────────────────

data "aws_ami" "ecs_optimized" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn2-ami-ecs-hvm-*-x86_64-ebs"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

# ── Launch Template ─────────────────────────────────────────────────────────

resource "aws_launch_template" "ecs_instance" {
  name_prefix   = "${local.name}-ecs-"
  image_id      = data.aws_ami.ecs_optimized.id
  instance_type = var.ecs_instance_type

  iam_instance_profile {
    arn = aws_iam_instance_profile.ecs_instance.arn
  }

  vpc_security_group_ids = [aws_security_group.ecs_instance.id]

  key_name = var.ecs_key_pair_name != "" ? var.ecs_key_pair_name : null

  monitoring {
    enabled = true
  }

  user_data = base64encode(<<-EOF
    #!/bin/bash
    echo ECS_CLUSTER=${aws_ecs_cluster.main.name} >> /etc/ecs/ecs.config
    echo ECS_ENABLE_CONTAINER_METADATA=true >> /etc/ecs/ecs.config
    echo ECS_ENABLE_TASK_IAM_ROLE=true >> /etc/ecs/ecs.config
    echo ECS_ENABLE_TASK_IAM_ROLE_NETWORK_HOST=true >> /etc/ecs/ecs.config
    
    # Install SSM agent for remote management
    yum install -y amazon-ssm-agent
    systemctl enable amazon-ssm-agent
    systemctl start amazon-ssm-agent
  EOF
  )

  tag_specifications {
    resource_type = "instance"
    tags = {
      Name = "${local.name}-ecs-instance"
    }
  }

  lifecycle {
    create_before_destroy = true
  }
}

# ── Security Group for EC2 Instances ────────────────────────────────────────

resource "aws_security_group" "ecs_instance" {
  name_prefix = "${local.name}-ecs-instance-"
  vpc_id      = aws_vpc.main.id
  description = "Security group for ECS EC2 instances"

  # Allow all traffic from ALB
  ingress {
    from_port       = 0
    to_port         = 65535
    protocol        = "tcp"
    security_groups = [aws_security_group.alb.id]
  }

  # Allow inter-instance communication (for service discovery)
  ingress {
    from_port = 0
    to_port   = 65535
    protocol  = "tcp"
    self      = true
  }

  # Allow all outbound traffic
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = { Name = "${local.name}-ecs-instance-sg" }

  lifecycle {
    create_before_destroy = true
  }
}

# ── Auto Scaling Group ──────────────────────────────────────────────────────

resource "aws_autoscaling_group" "ecs" {
  name                      = "${local.name}-ecs-asg"
  vpc_zone_identifier       = aws_subnet.private[*].id
  min_size                  = var.ecs_asg_min_size
  max_size                  = var.ecs_asg_max_size
  desired_capacity          = var.ecs_asg_desired_capacity
  health_check_type         = "EC2"
  health_check_grace_period = 300
  protect_from_scale_in     = true # Managed by ECS capacity provider

  launch_template {
    id      = aws_launch_template.ecs_instance.id
    version = "$Latest"
  }

  tag {
    key                 = "Name"
    value               = "${local.name}-ecs-instance"
    propagate_at_launch = true
  }

  tag {
    key                 = "AmazonECSManaged"
    value               = "true"
    propagate_at_launch = true
  }

  lifecycle {
    create_before_destroy = true
    ignore_changes        = [desired_capacity]
  }
}

# ── Auto Scaling Policies ───────────────────────────────────────────────────

resource "aws_autoscaling_policy" "ecs_scale_up" {
  name                   = "${local.name}-ecs-scale-up"
  autoscaling_group_name = aws_autoscaling_group.ecs.name
  adjustment_type        = "ChangeInCapacity"
  scaling_adjustment     = 1
  cooldown               = 300
}

resource "aws_autoscaling_policy" "ecs_scale_down" {
  name                   = "${local.name}-ecs-scale-down"
  autoscaling_group_name = aws_autoscaling_group.ecs.name
  adjustment_type        = "ChangeInCapacity"
  scaling_adjustment     = -1
  cooldown               = 300
}

# ── CloudWatch Alarms for ASG Scaling ───────────────────────────────────────

resource "aws_cloudwatch_metric_alarm" "ecs_cpu_high" {
  alarm_name          = "${local.name}-ecs-cpu-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "CPUReservation"
  namespace           = "AWS/ECS"
  period              = 60
  statistic           = "Average"
  threshold           = 70
  alarm_description   = "Triggers when ECS cluster CPU reservation is high"
  alarm_actions       = [aws_autoscaling_policy.ecs_scale_up.arn]

  dimensions = {
    ClusterName = aws_ecs_cluster.main.name
  }
}

resource "aws_cloudwatch_metric_alarm" "ecs_cpu_low" {
  alarm_name          = "${local.name}-ecs-cpu-low"
  comparison_operator = "LessThanThreshold"
  evaluation_periods  = 3
  metric_name         = "CPUReservation"
  namespace           = "AWS/ECS"
  period              = 60
  statistic           = "Average"
  threshold           = 30
  alarm_description   = "Triggers when ECS cluster CPU reservation is low"
  alarm_actions       = [aws_autoscaling_policy.ecs_scale_down.arn]

  dimensions = {
    ClusterName = aws_ecs_cluster.main.name
  }
}

resource "aws_cloudwatch_metric_alarm" "ecs_memory_high" {
  alarm_name          = "${local.name}-ecs-memory-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "MemoryReservation"
  namespace           = "AWS/ECS"
  period              = 60
  statistic           = "Average"
  threshold           = 70
  alarm_description   = "Triggers when ECS cluster memory reservation is high"
  alarm_actions       = [aws_autoscaling_policy.ecs_scale_up.arn]

  dimensions = {
    ClusterName = aws_ecs_cluster.main.name
  }
}
