export const siteConfig = {
  name: "MyPaaS",
  description: "The ultimate platform for your SaaS applications.",
  nav: [
    { label: "Features", href: "#features" },
    { label: "Pricing", href: "#pricing" },
    { label: "About", href: "#about" },
  ],
  hero: {
    title: "Deploy your App in Seconds",
    subtitle: "Focus on your code, we handle the infrastructure. Scalable, secure, and developer-friendly.",
    cta: "Get Started",
    image: "/hero-image.png"
  },
  features: [
    {
      title: "Auto-Scaling",
      description: "Handle traffic spikes effortlessly with our auto-scaling infrastructure.",
      icon: "Activity"
    },
    {
      title: "Global Edge",
      description: "Deploy your app to 35+ regions worldwide for low latency.",
      icon: "Globe"
    },
    {
      title: "Zero Config",
      description: "Just push your code and we handle the rest. No YAML required.",
      icon: "Zap"
    },
    {
      title: "Secure by Design",
      description: "Enterprise-grade security with automated patching and compliance.",
      icon: "Shield"
    }
  ],
  pricing: [
    {
      name: "Hobby",
      price: "$0",
      features: ["1 Project", "Community Support", "Shared Resources"],
      cta: "Start Free"
    },
    {
      name: "Pro",
      price: "$29",
      features: ["Unlimited Projects", "Priority Support", "Dedicated Resources"],
      cta: "Go Pro",
      popular: true
    },
    {
      name: "Enterprise",
      price: "Custom",
      features: ["SLA", "Account Manager", "On-premise Deployment"],
      cta: "Contact Us"
    }
  ]
};
