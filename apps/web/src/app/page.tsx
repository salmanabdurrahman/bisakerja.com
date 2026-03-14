import { AppShell } from "@/components/layout/app-shell";
import { ButtonLink } from "@/components/ui/button";
import {
  HiMagnifyingGlass,
  HiArrowTrendingUp,
  HiBolt,
  HiUserGroup,
  HiBriefcase,
  HiGlobeAlt,
  HiCheckCircle,
  HiDocumentText,
  HiBell,
  HiStar,
} from "react-icons/hi2";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

export default function Home() {
  return (
    <AppShell>
      <main className="grid gap-20 sm:gap-24 pb-16" role="main">
        {/* Hero Section */}
        <section className="flex flex-col items-center text-center px-4 pt-12 sm:pt-20 md:pt-24 pb-12 sm:pb-24">
          <h1 className="max-w-4xl text-[36px] sm:text-[48px] md:text-[64px] lg:text-[72px] font-normal leading-[1.1] tracking-tight text-black font-display">
            The professional{" "}
            <span className="italic font-light">workspace</span> for job
            discovery.
          </h1>

          <p className="max-w-150 mt-6 text-[16px] sm:text-[18px] font-normal leading-relaxed text-[#666666]">
            Filter by role, save high-signal searches, and receive real-time
            premium alerts. Your next career move, simplified.
          </p>

          <div className="mt-10 flex flex-col sm:flex-row w-full sm:w-auto justify-center gap-4">
            <ButtonLink
              href="/auth/register"
              size="lg"
              variant="primary"
              className="w-full sm:w-auto sm:min-w-40"
            >
              Start for free
            </ButtonLink>
            <ButtonLink
              href="/pricing"
              size="lg"
              variant="outline"
              className="w-full sm:w-auto sm:min-w-40"
            >
              Book a demo
            </ButtonLink>
          </div>
        </section>

        {/* Hero Image Section */}
        <section className="mx-auto w-full max-w-5xl rounded-4xl sm:rounded-[48px] bg-[#F4F4F4] p-4 sm:p-6 mb-12">
          <div className="relative w-full aspect-video md:aspect-21/9 rounded-3xl bg-white border border-[#E5E5E5] shadow-sm flex flex-col overflow-hidden">
            <div className="flex items-center justify-between px-6 py-4 border-b border-[#E5E5E5] bg-[#F9F9F9]">
              <div className="flex items-center gap-2">
                <div className="h-3 w-3 rounded-full bg-[#E5E5E5]"></div>
                <div className="h-3 w-3 rounded-full bg-[#E5E5E5]"></div>
                <div className="h-3 w-3 rounded-full bg-[#E5E5E5]"></div>
              </div>
              <div className="h-6 w-32 rounded bg-[#E5E5E5]/50 hidden sm:block"></div>
            </div>
            <div className="p-6 md:p-8 grid gap-6 bg-white flex-1">
              <div className="flex flex-col sm:flex-row gap-4 justify-between">
                <div className="h-8 w-1/3 rounded-md bg-[#F4F4F4]"></div>
                <div className="h-8 w-1/4 rounded-md bg-[#F4F4F4] hidden sm:block"></div>
              </div>
              <div className="grid gap-4">
                {[...Array(3)].map((_, i) => (
                  <div
                    key={i}
                    className="h-16 md:h-20 w-full rounded-xl bg-white border border-[#E5E5E5] flex items-center px-4 gap-4 transition-all duration-200 hover:border-black/20"
                  >
                    <div className="h-8 w-8 md:h-10 md:w-10 rounded bg-[#F4F4F4]"></div>
                    <div className="flex-1 grid gap-2">
                      <div className="h-3 md:h-4 w-1/3 rounded bg-[#F4F4F4]"></div>
                      <div className="h-2 md:h-3 w-1/4 rounded bg-[#F4F4F4]"></div>
                    </div>
                    <div className="h-6 md:h-8 w-20 md:w-24 rounded-full bg-[#F9F9F9] hidden sm:block"></div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </section>

        {/* Features Section */}
        <section className="grid gap-6 pt-12">
          <div className="text-center max-w-2xl mx-auto mb-12">
            <h2 className="text-[40px] sm:text-[48px] font-normal tracking-tight text-black font-display leading-[1.1]">
              Everything you need to move faster
            </h2>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-6 px-4">
            <Card className="flex flex-col p-6 transition-all duration-200 hover:-translate-y-1 hover:shadow-md">
              <CardHeader className="p-0 mb-4">
                <div className="h-12 w-12 rounded-full bg-[#F4F4F4] flex items-center justify-center mb-4">
                  <HiMagnifyingGlass className="h-6 w-6 text-black" />
                </div>
                <CardTitle className="text-xl sm:text-2xl font-normal">
                  Search with precision
                </CardTitle>
              </CardHeader>
              <CardContent className="p-0">
                <p className="text-[14px] text-[#666666] font-normal leading-relaxed">
                  Filter by role, location, salary range, and source in one
                  clean URL-driven workflow.
                </p>
              </CardContent>
            </Card>

            <Card className="flex flex-col p-6 transition-all duration-200 hover:-translate-y-1 hover:shadow-md">
              <CardHeader className="p-0 mb-4">
                <div className="h-12 w-12 rounded-full bg-[#F4F4F4] flex items-center justify-center mb-4">
                  <HiArrowTrendingUp className="h-6 w-6 text-black" />
                </div>
                <CardTitle className="text-xl sm:text-2xl font-normal">
                  Track growth signals
                </CardTitle>
              </CardHeader>
              <CardContent className="p-0">
                <p className="text-[14px] text-[#666666] font-normal leading-relaxed">
                  Save high-signal searches, tune digest cadence, and keep your
                  opportunity pipeline active.
                </p>
              </CardContent>
            </Card>

            <Card className="flex flex-col p-6 transition-all duration-200 hover:-translate-y-1 hover:shadow-md">
              <CardHeader className="p-0 mb-4">
                <div className="h-12 w-12 rounded-full bg-[#F4F4F4] flex items-center justify-center mb-4">
                  <HiBolt className="h-6 w-6 text-black" />
                </div>
                <CardTitle className="text-xl sm:text-2xl font-normal">
                  Upgrade effortlessly
                </CardTitle>
              </CardHeader>
              <CardContent className="p-0">
                <p className="text-[14px] text-[#666666] font-normal leading-relaxed">
                  Start checkout in seconds, monitor premium status, and stay
                  ahead of the competition.
                </p>
              </CardContent>
            </Card>
          </div>
        </section>

        {/* Stats Section */}
        <section className="py-16 sm:py-20 bg-[#F4F4F4] mx-4 sm:mx-8 rounded-4xl sm:rounded-[48px]">
          <div className="max-w-6xl mx-auto px-6 sm:px-12">
            <div className="grid grid-cols-2 md:grid-cols-4 gap-8 md:gap-12">
              <div className="text-center">
                <div className="text-[36px] sm:text-[48px] font-normal text-black tracking-tight">
                  50K+
                </div>
                <p className="text-[14px] sm:text-[16px] text-[#666666] mt-2">
                  Active job listings
                </p>
              </div>
              <div className="text-center">
                <div className="text-[36px] sm:text-[48px] font-normal text-black tracking-tight">
                  10K+
                </div>
                <p className="text-[14px] sm:text-[16px] text-[#666666] mt-2">
                  Companies hiring
                </p>
              </div>
              <div className="text-center">
                <div className="text-[36px] sm:text-[48px] font-normal text-black tracking-tight">
                  3
                </div>
                <p className="text-[14px] sm:text-[16px] text-[#666666] mt-2">
                  Job sources integrated
                </p>
              </div>
              <div className="text-center">
                <div className="text-[36px] sm:text-[48px] font-normal text-black tracking-tight">
                  24/7
                </div>
                <p className="text-[14px] sm:text-[16px] text-[#666666] mt-2">
                  Real-time updates
                </p>
              </div>
            </div>
          </div>
        </section>

        {/* How It Works Section */}
        <section className="py-16 sm:py-20 px-4">
          <div className="text-center max-w-2xl mx-auto mb-16">
            <h2 className="text-[40px] sm:text-[48px] font-normal tracking-tight text-black font-display leading-[1.1]">
              How it works
            </h2>
            <p className="text-[16px] sm:text-[18px] text-[#666666] mt-4">
              Get started in minutes with our simple three-step process
            </p>
          </div>

          <div className="max-w-5xl mx-auto">
            <div className="grid md:grid-cols-3 gap-8 md:gap-6">
              <div className="relative flex flex-col items-center text-center p-6">
                <div className="h-16 w-16 rounded-full bg-black text-white flex items-center justify-center mb-6 text-[24px] font-normal">
                  1
                </div>
                <div className="h-14 w-14 rounded-full bg-[#F4F4F4] flex items-center justify-center mb-4">
                  <HiDocumentText className="h-7 w-7 text-black" />
                </div>
                <h3 className="text-xl sm:text-2xl font-normal mb-3">
                  Create your profile
                </h3>
                <p className="text-[14px] text-[#666666] leading-relaxed">
                  Sign up for free and set your job preferences — role,
                  location, salary range, and more.
                </p>
                <div className="hidden md:block absolute top-12 -right-3 w-6 h-0.5 bg-[#E5E5E5]"></div>
              </div>

              <div className="relative flex flex-col items-center text-center p-6">
                <div className="h-16 w-16 rounded-full bg-black text-white flex items-center justify-center mb-6 text-[24px] font-normal">
                  2
                </div>
                <div className="h-14 w-14 rounded-full bg-[#F4F4F4] flex items-center justify-center mb-4">
                  <HiMagnifyingGlass className="h-7 w-7 text-black" />
                </div>
                <h3 className="text-xl sm:text-2xl font-normal mb-3">
                  Search & save
                </h3>
                <p className="text-[14px] text-[#666666] leading-relaxed">
                  Browse thousands of curated jobs from top sources. Save
                  searches and get instant notifications.
                </p>
                <div className="hidden md:block absolute top-12 -right-3 w-6 h-0.5 bg-[#E5E5E5]"></div>
              </div>

              <div className="flex flex-col items-center text-center p-6">
                <div className="h-16 w-16 rounded-full bg-black text-white flex items-center justify-center mb-6 text-[24px] font-normal">
                  3
                </div>
                <div className="h-14 w-14 rounded-full bg-[#F4F4F4] flex items-center justify-center mb-4">
                  <HiBell className="h-7 w-7 text-black" />
                </div>
                <h3 className="text-xl sm:text-2xl font-normal mb-3">
                  Get notified
                </h3>
                <p className="text-[14px] text-[#666666] leading-relaxed">
                  Receive real-time alerts for new matches. Upgrade to premium
                  for priority notifications.
                </p>
              </div>
            </div>
          </div>
        </section>

        {/* Job Sources Section */}
        <section className="py-16 sm:py-20 px-4">
          <div className="text-center max-w-2xl mx-auto mb-12">
            <h2 className="text-[40px] sm:text-[48px] font-normal tracking-tight text-black font-display leading-[1.1]">
              Aggregated from top sources
            </h2>
            <p className="text-[16px] sm:text-[18px] text-[#666666] mt-4">
              We collect and curate job listings from Indonesia&apos;s leading
              job platforms
            </p>
          </div>

          <div className="max-w-4xl mx-auto">
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-6">
              <div className="bk-card p-8 flex flex-col items-center text-center transition-all duration-200 hover:-translate-y-1 hover:shadow-md">
                <div className="h-16 w-16 rounded-2xl bg-[#00A19C]/10 flex items-center justify-center mb-4">
                  <HiBriefcase className="h-8 w-8 text-[#00A19C]" />
                </div>
                <h3 className="text-xl font-normal mb-2">Glints</h3>
                <p className="text-[14px] text-[#666666]">
                  Southeast Asia&apos;s leading career platform
                </p>
              </div>

              <div className="bk-card p-8 flex flex-col items-center text-center transition-all duration-200 hover:-translate-y-1 hover:shadow-md">
                <div className="h-16 w-16 rounded-2xl bg-[#3B82F6]/10 flex items-center justify-center mb-4">
                  <HiUserGroup className="h-8 w-8 text-[#3B82F6]" />
                </div>
                <h3 className="text-xl font-normal mb-2">Kalibrr</h3>
                <p className="text-[14px] text-[#666666]">
                  AI-powered job matching platform
                </p>
              </div>

              <div className="bk-card p-8 flex flex-col items-center text-center transition-all duration-200 hover:-translate-y-1 hover:shadow-md">
                <div className="h-16 w-16 rounded-2xl bg-[#6366F1]/10 flex items-center justify-center mb-4">
                  <HiGlobeAlt className="h-8 w-8 text-[#6366F1]" />
                </div>
                <h3 className="text-xl font-normal mb-2">JobStreet</h3>
                <p className="text-[14px] text-[#666666]">
                  Asia&apos;s premier job portal
                </p>
              </div>
            </div>
          </div>
        </section>

        {/* Testimonials Section */}
        <section className="py-16 sm:py-20 bg-[#F4F4F4] mx-4 sm:mx-8 rounded-4xl sm:rounded-[48px]">
          <div className="max-w-6xl mx-auto px-6 sm:px-12">
            <div className="text-center max-w-2xl mx-auto mb-16">
              <h2 className="text-[40px] sm:text-[48px] font-normal tracking-tight text-black font-display leading-[1.1]">
                Loved by job seekers
              </h2>
              <p className="text-[16px] sm:text-[18px] text-[#666666] mt-4">
                See what professionals are saying about BisaKerja
              </p>
            </div>

            <div className="grid md:grid-cols-3 gap-6">
              <div className="bg-white rounded-3xl p-8 shadow-sm">
                <div className="flex gap-1 mb-4">
                  {[...Array(5)].map((_, i) => (
                    <HiStar
                      key={i}
                      className="h-5 w-5 text-yellow-400 fill-yellow-400"
                    />
                  ))}
                </div>
                <p className="text-[14px] sm:text-[16px] text-[#333333] leading-relaxed mb-6">
                  &quot;Finally, a job search tool that actually saves me time.
                  The filters are precise and the alerts are on point. Found my
                  dream job in just 2 weeks!&quot;
                </p>
                <div className="flex items-center gap-3">
                  <div className="h-12 w-12 rounded-full bg-[#F4F4F4] flex items-center justify-center">
                    <span className="text-[16px] font-medium">RA</span>
                  </div>
                  <div>
                    <p className="font-medium text-black">Rizky Adiputra</p>
                    <p className="text-[14px] text-[#666666]">
                      Software Engineer
                    </p>
                  </div>
                </div>
              </div>

              <div className="bg-white rounded-3xl p-8 shadow-sm">
                <div className="flex gap-1 mb-4">
                  {[...Array(5)].map((_, i) => (
                    <HiStar
                      key={i}
                      className="h-5 w-5 text-yellow-400 fill-yellow-400"
                    />
                  ))}
                </div>
                <p className="text-[14px] sm:text-[16px] text-[#333333] leading-relaxed mb-6">
                  &quot;The aggregated listings save hours of browsing multiple
                  sites. Premium alerts helped me catch opportunities before
                  everyone else.&quot;
                </p>
                <div className="flex items-center gap-3">
                  <div className="h-12 w-12 rounded-full bg-[#F4F4F4] flex items-center justify-center">
                    <span className="text-[16px] font-medium">SP</span>
                  </div>
                  <div>
                    <p className="font-medium text-black">Sarah Putri</p>
                    <p className="text-[14px] text-[#666666]">
                      Product Manager
                    </p>
                  </div>
                </div>
              </div>

              <div className="bg-white rounded-3xl p-8 shadow-sm">
                <div className="flex gap-1 mb-4">
                  {[...Array(5)].map((_, i) => (
                    <HiStar
                      key={i}
                      className="h-5 w-5 text-yellow-400 fill-yellow-400"
                    />
                  ))}
                </div>
                <p className="text-[14px] sm:text-[16px] text-[#333333] leading-relaxed mb-6">
                  &quot;Clean interface, fast search, real-time updates.
                  BisaKerja understands what modern job seekers need. Highly
                  recommended!&quot;
                </p>
                <div className="flex items-center gap-3">
                  <div className="h-12 w-12 rounded-full bg-[#F4F4F4] flex items-center justify-center">
                    <span className="text-[16px] font-medium">AW</span>
                  </div>
                  <div>
                    <p className="font-medium text-black">Andi Wijaya</p>
                    <p className="text-[14px] text-[#666666]">Data Analyst</p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>

        {/* Free vs Premium Section */}
        <section className="py-16 sm:py-20 px-4">
          <div className="text-center max-w-2xl mx-auto mb-12">
            <h2 className="text-[40px] sm:text-[48px] font-normal tracking-tight text-black font-display leading-[1.1]">
              Choose your plan
            </h2>
            <p className="text-[16px] sm:text-[18px] text-[#666666] mt-4">
              Start free, upgrade when you need more power
            </p>
          </div>

          <div className="max-w-4xl mx-auto">
            <div className="grid md:grid-cols-2 gap-6">
              <div className="bk-card p-8">
                <h3 className="text-2xl font-normal mb-2">Free</h3>
                <p className="text-[14px] text-[#666666] mb-6">
                  Perfect for getting started
                </p>
                <div className="text-[40px] font-normal mb-8">
                  Rp 0
                  <span className="text-[16px] text-[#666666] font-normal">
                    /month
                  </span>
                </div>
                <ul className="grid gap-4 mb-8">
                  <li className="flex items-center gap-3 text-[14px]">
                    <HiCheckCircle className="h-5 w-5 text-emerald-500 shrink-0" />
                    <span>Browse all job listings</span>
                  </li>
                  <li className="flex items-center gap-3 text-[14px]">
                    <HiCheckCircle className="h-5 w-5 text-emerald-500 shrink-0" />
                    <span>Basic search filters</span>
                  </li>
                  <li className="flex items-center gap-3 text-[14px]">
                    <HiCheckCircle className="h-5 w-5 text-emerald-500 shrink-0" />
                    <span>Save up to 3 searches</span>
                  </li>
                  <li className="flex items-center gap-3 text-[14px]">
                    <HiCheckCircle className="h-5 w-5 text-emerald-500 shrink-0" />
                    <span>Weekly email digest</span>
                  </li>
                </ul>
                <ButtonLink
                  href="/auth/register"
                  variant="outline"
                  size="lg"
                  fullWidth
                >
                  Get started free
                </ButtonLink>
              </div>

              <div className="bk-card p-8 border-2 border-black relative">
                <div className="absolute -top-3 left-8 bg-black text-white text-[12px] px-3 py-1 rounded-full">
                  Most popular
                </div>
                <h3 className="text-2xl font-normal mb-2">Premium</h3>
                <p className="text-[14px] text-[#666666] mb-6">
                  For serious job seekers
                </p>
                <div className="text-[40px] font-normal mb-8">
                  Rp 49K
                  <span className="text-[16px] text-[#666666] font-normal">
                    /month
                  </span>
                </div>
                <ul className="grid gap-4 mb-8">
                  <li className="flex items-center gap-3 text-[14px]">
                    <HiCheckCircle className="h-5 w-5 text-emerald-500 shrink-0" />
                    <span>Everything in Free</span>
                  </li>
                  <li className="flex items-center gap-3 text-[14px]">
                    <HiCheckCircle className="h-5 w-5 text-emerald-500 shrink-0" />
                    <span>Advanced salary & location filters</span>
                  </li>
                  <li className="flex items-center gap-3 text-[14px]">
                    <HiCheckCircle className="h-5 w-5 text-emerald-500 shrink-0" />
                    <span>Unlimited saved searches</span>
                  </li>
                  <li className="flex items-center gap-3 text-[14px]">
                    <HiCheckCircle className="h-5 w-5 text-emerald-500 shrink-0" />
                    <span>Real-time job alerts</span>
                  </li>
                  <li className="flex items-center gap-3 text-[14px]">
                    <HiCheckCircle className="h-5 w-5 text-emerald-500 shrink-0" />
                    <span>Priority support</span>
                  </li>
                </ul>
                <ButtonLink
                  href="/pricing"
                  variant="primary"
                  size="lg"
                  fullWidth
                >
                  Upgrade to Premium
                </ButtonLink>
              </div>
            </div>
          </div>
        </section>
        {/* CTA Section */}
        <section className="flex flex-col items-center justify-center text-center mt-24 mb-12 bg-[#1A1A1A] mx-4 sm:mx-8 rounded-4xl sm:rounded-[48px] py-16 sm:py-20 px-6 sm:px-12">
          <h2 className="text-[32px] sm:text-[40px] md:text-[48px] font-normal tracking-tight font-display mb-6 text-white leading-[1.1]">
            Ready to find your next role?
          </h2>
          <p className="text-[16px] sm:text-[18px] text-[#A0A0A0] font-normal max-w-150 mx-auto mb-10">
            Join thousands of professionals who have upgraded their job search
            experience. Start for free and upgrade when you need more power.
          </p>
          <div className="flex flex-col sm:flex-row w-full sm:w-auto justify-center gap-4">
            <ButtonLink
              href="/auth/register"
              size="lg"
              className="w-full sm:w-auto sm:min-w-40 bg-white text-black hover:bg-gray-200"
            >
              Create Account
            </ButtonLink>
            <ButtonLink
              href="/pricing"
              variant="outline"
              size="lg"
              className="w-full sm:w-auto sm:min-w-40 border-[#333333] text-white hover:bg-[#333333] hover:text-black"
            >
              View Pricing
            </ButtonLink>
          </div>
        </section>
      </main>
    </AppShell>
  );
}
