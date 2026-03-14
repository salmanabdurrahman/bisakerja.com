import { AppShell } from "@/components/layout/app-shell";
import { ButtonLink } from "@/components/ui/button";
import { HiMagnifyingGlass, HiArrowTrendingUp, HiBolt } from "react-icons/hi2";
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
